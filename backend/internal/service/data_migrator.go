package service

import (
	"context"
	"fmt"
	"log"

	"gridea-pro/backend/internal/domain"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// DataMigrator 负责在应用启动时全量清洗和迁移本地老旧格式的 ID
// 把不合规的分类和标签 ID 洗刷为统一的 6位 NanoID，并将文章（Post）中针对名字/别名的关联
// 统一投射、补足为符合当前字典映射的标准 CategoryIDs / TagIDs 数组。
type DataMigrator struct {
	appDir       string
	categoryRepo domain.CategoryRepository
	tagRepo      domain.TagRepository
	postRepo     domain.PostRepository
	menuRepo     domain.MenuRepository
	linkRepo     domain.LinkRepository
	memoRepo     domain.MemoRepository

	// NanoID 统一生成规范
	alphabet string
	length   int
}

func NewDataMigrator(
	appDir string,
	categoryRepo domain.CategoryRepository,
	tagRepo domain.TagRepository,
	postRepo domain.PostRepository,
	menuRepo domain.MenuRepository,
	linkRepo domain.LinkRepository,
	memoRepo domain.MemoRepository,
) *DataMigrator {
	return &DataMigrator{
		appDir:       appDir,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		postRepo:     postRepo,
		menuRepo:     menuRepo,
		linkRepo:     linkRepo,
		memoRepo:     memoRepo,

		// 全局约束
		alphabet: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		length:   6,
	}
}

// generateID 内部统一下发 ID
func (m *DataMigrator) generateID() string {
	id, _ := gonanoid.Generate(m.alphabet, m.length)
	return id
}

// isValidID 检查当前 ID 是否符合 6位长度 以及纯粹的字母表规范
func (m *DataMigrator) isValidID(id string) bool {
	if len(id) != m.length {
		return false
	}
	// 简单校验是否都在 alphabet 内（可选，通常长度不对就足以判断是否是用老的方式例如 UUID 或 9 位等生成的）
	for _, char := range id {
		isValidChar := false
		for _, validChar := range m.alphabet {
			if char == validChar {
				isValidChar = true
				break
			}
		}
		if !isValidChar {
			return false
		}
	}
	return true
}

func (m *DataMigrator) RunMigration(ctx context.Context) error {
	log.Println("[DataMigrator] --------- 开始全量检查与迁移历史基础关联数据 ID ---------")

	// ---------------- 第一步：基础数据清洗与映射构建 ----------------

	// 1.1 获取并洗刷分类 (Category)
	categories, err := m.categoryRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("加载分类失败: %w", err)
	}

	categorySlugToIDMap := make(map[string]string) // 老 Slug -> 新 ID
	categoryNameToIDMap := make(map[string]string) // 老 Name -> 新 ID (作为备用冗余)
	var categoryNeedsSave bool

	for i, cat := range categories {
		oldID := cat.ID
		if !m.isValidID(cat.ID) { // ID为空或长度不对
			newID := m.generateID()
			categories[i].ID = newID
			categoryNeedsSave = true
			log.Printf("[DataMigrator] 分类 [%s] ID 不合规或为空 (%s) -> 分配新 ID: %s", cat.Name, oldID, newID)
		}
		// 加入映射大表，无论原本是否合规，都需入表，方便供 Post 检索引用
		categorySlugToIDMap[categories[i].Slug] = categories[i].ID
		categoryNameToIDMap[categories[i].Name] = categories[i].ID
	}

	if categoryNeedsSave {
		// 回写 categories.json
		// 因为很多老数据原来是没有 ID 的，如果调 Update 会报 "item not found"。所以必须用 SaveAll 全量覆盖保存。
		if err := m.categoryRepo.SaveAll(ctx, categories); err != nil {
			log.Printf("[DataMigrator] 保存修复后的分类数据失败: %v\n", err)
		} else {
			log.Println("[DataMigrator] 成功保存修复后的分类数据。")
		}
	}

	// 1.2 获取并洗刷标签 (Tag)
	tags, err := m.tagRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("加载标签失败: %w", err)
	}

	tagNameToIDMap := make(map[string]string)
	var tagNeedsSave bool

	for i, tag := range tags {
		oldID := tag.ID
		if !m.isValidID(tag.ID) {
			newID := m.generateID()
			tags[i].ID = newID
			tagNeedsSave = true
			log.Printf("[DataMigrator] 标签 [%s] ID 不合规或为空 (%s) -> 分配新 ID: %s", tag.Name, oldID, newID)
		}
		tagNameToIDMap[tags[i].Name] = tags[i].ID
	}

	if tagNeedsSave {
		if err := m.tagRepo.SaveAll(ctx, tags); err != nil {
			log.Printf("[DataMigrator] 保存修复后的标签数据失败: %v\n", err)
		} else {
			log.Println("[DataMigrator] 成功保存修复后的标签数据。")
		}
	}

	// 1.3 获取并洗刷菜单 (Menu)
	menus, err := m.menuRepo.List(ctx)
	if err == nil {
		var menuNeedsSave bool
		for i, menu := range menus {
			if !m.isValidID(menu.ID) {
				newID := m.generateID()
				menus[i].ID = newID
				menuNeedsSave = true
				log.Printf("[DataMigrator] 菜单 [%s] ID 不合规或为空 -> 分配新 ID: %s", menu.Name, newID)
			}
		}
		if menuNeedsSave {
			if err := m.menuRepo.SaveAll(ctx, menus); err != nil {
				log.Printf("[DataMigrator] 保存修复后的菜单数据失败: %v\n", err)
			} else {
				log.Println("[DataMigrator] 成功保存修复后的菜单数据。")
			}
		}
	} else {
		log.Printf("[DataMigrator] 加载菜单失败 (跳过): %v", err)
	}

	// 1.4 获取并洗刷友链 (Link)
	links, err := m.linkRepo.List(ctx)
	if err == nil {
		var linkNeedsSave bool
		for i, link := range links {
			if !m.isValidID(link.ID) {
				newID := m.generateID()
				links[i].ID = newID
				linkNeedsSave = true
				log.Printf("[DataMigrator] 友链 [%s] ID 不合规或为空 -> 分配新 ID: %s", link.Name, newID)
			}
		}
		if linkNeedsSave {
			if err := m.linkRepo.SaveAll(ctx, links); err != nil {
				log.Printf("[DataMigrator] 保存修复后的友链数据失败: %v\n", err)
			} else {
				log.Println("[DataMigrator] 成功保存修复后的友链数据。")
			}
		}
	} else {
		log.Printf("[DataMigrator] 加载友链失败 (跳过): %v", err)
	}

	// 1.5 获取并洗刷闪念 (Memo)
	memos, err := m.memoRepo.List(ctx)
	if err == nil {
		var memoNeedsSave bool
		for i, memo := range memos {
			if !m.isValidID(memo.ID) {
				newID := m.generateID()
				memos[i].ID = newID
				memoNeedsSave = true
				log.Printf("[DataMigrator] 闪念 [...] ID 不合规或为空 -> 分配新 ID: %s", newID)
			}
		}
		if memoNeedsSave {
			if err := m.memoRepo.SaveAll(ctx, memos); err != nil {
				log.Printf("[DataMigrator] 保存修复后的闪念快照数据失败: %v\n", err)
			} else {
				log.Println("[DataMigrator] 成功保存修复后的闪念数据。")
			}
		}
	} else {
		log.Printf("[DataMigrator] 加载闪念失败 (跳过): %v", err)
	}

	// ---------------- 第二步：文章关联关系的修复 ----------------

	posts, err := m.postRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("加载文章聚合失败: %w", err)
	}

	var migratedPostCount int

	for _, post := range posts {
		var postModified bool

		// 2.0 修复 Post 自带的 ID
		oldPostID := post.ID
		if !m.isValidID(post.ID) {
			newID := m.generateID()
			post.ID = newID
			postModified = true
			log.Printf("[DataMigrator] 文章 [%s] ID 不合规或为空 (%s) -> 分配新 ID: %s", post.Title, oldPostID, newID)
		}

		// 2.1 修复 CategoryIDs
		// 如果文章的老 Categories 字段有值，但 CategoryIDs 映射数量对不上，或者发现残存的旧格式数据
		if len(post.Categories) > 0 {
			var newCategoryIDs []string
			for _, catIdent := range post.Categories {
				// 首先尝试当匹配 Slug
				if mappedID, ok := categorySlugToIDMap[catIdent]; ok {
					newCategoryIDs = append(newCategoryIDs, mappedID)
					continue
				}
				// 尝试匹配 Name
				if mappedID, ok := categoryNameToIDMap[catIdent]; ok {
					newCategoryIDs = append(newCategoryIDs, mappedID)
					continue
				}
				// 如果原本就是存的 ID 的情况
				if m.isValidID(catIdent) {
					newCategoryIDs = append(newCategoryIDs, catIdent)
					continue
				}
			}

			// 只有发生了变化才覆写 (为了防抖)，这可以通过简单的数量与元素对比得出
			if !slicesEqual(post.CategoryIDs, newCategoryIDs) {
				post.CategoryIDs = newCategoryIDs
				postModified = true
			}
		}

		// 2.2 修复 TagIDs
		if len(post.Tags) > 0 {
			var newTagIDs []string
			for _, tagName := range post.Tags {
				if mappedID, ok := tagNameToIDMap[tagName]; ok {
					newTagIDs = append(newTagIDs, mappedID)
				} else if m.isValidID(tagName) {
					// 兼容部分人直接把新 ID 填在 tags[] 里
					newTagIDs = append(newTagIDs, tagName)
				}
			}

			if !slicesEqual(post.TagIDs, newTagIDs) {
				post.TagIDs = newTagIDs
				postModified = true
			}
		}

		if postModified {
			if err := m.postRepo.Update(ctx, &post); err != nil {
				log.Printf("[DataMigrator] 回写修复文章失败 [%s]: %v\n", post.FileName, err)
			} else {
				migratedPostCount++
			}
		}
	}

	if migratedPostCount > 0 {
		log.Printf("[DataMigrator] 成功修复了 %d 篇文章的底层 ID 与映射关联关系。\n", migratedPostCount)
	} else {
		log.Println("[DataMigrator] ✅ 检查完成，未发现需要修正的历史数据，所有数据均处于最新健康状态。")
	}

	log.Println("[DataMigrator] --------- 数据清洗迁移协程运行完毕 ---------")
	return nil
}

// slicesEqual 比较两个 string 切片内容是否一致（无序集合比较，解决顺序不同导致的反复重写）
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// 如果都是空，也是相等
	if len(a) == 0 {
		return true
	}

	countMap := make(map[string]int)
	for _, item := range a {
		countMap[item]++
	}

	for _, item := range b {
		countMap[item]--
		if countMap[item] < 0 {
			return false
		}
	}

	// 此时所有的 count 应该都降回 0 了，因为长度相等且没出现负数
	return true
}
