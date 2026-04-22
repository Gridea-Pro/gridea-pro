package repository

import (
	"context"
	"errors"
	"fmt"

	"gridea-pro/backend/internal/domain"
)

// 数据层唯一性错误：由 tag/category Create/Update 返回，service 层可以用 errors.Is 判断。
var (
	ErrDuplicateName = errors.New("name already exists")
	ErrDuplicateSlug = errors.New("slug already exists")
)

// checkTagUniqueness 在现有 list 中查找与 target 冲突的记录，excludeID 指定要忽略的自身 ID。
// 错误消息面向用户，包装 ErrDuplicateName / ErrDuplicateSlug 以便 service 层用 errors.Is 识别类型。
func checkTagUniqueness(list []domain.Tag, target domain.Tag, excludeID string) error {
	for _, t := range list {
		if t.ID == excludeID {
			continue
		}
		if target.Name != "" && t.Name == target.Name {
			return fmt.Errorf("%w：标签名 %q 已被使用", ErrDuplicateName, target.Name)
		}
		if target.Slug != "" && t.Slug == target.Slug {
			return fmt.Errorf("%w：标签 URL slug %q 已被使用", ErrDuplicateSlug, target.Slug)
		}
	}
	return nil
}

// checkCategoryUniqueness 同上，针对 Category。
func checkCategoryUniqueness(list []domain.Category, target domain.Category, excludeID string) error {
	for _, c := range list {
		if c.ID == excludeID {
			continue
		}
		if target.Name != "" && c.Name == target.Name {
			return fmt.Errorf("%w：分类名 %q 已被使用", ErrDuplicateName, target.Name)
		}
		if target.Slug != "" && c.Slug == target.Slug {
			return fmt.Errorf("%w：分类 URL slug %q 已被使用", ErrDuplicateSlug, target.Slug)
		}
	}
	return nil
}

// AuditTagUniqueness 扫描仓库中所有标签，返回可读的 Name / Slug 冲突描述。
// 用于应用启动时检测手工编辑 JSON 带来的历史重复数据。
func AuditTagUniqueness(ctx context.Context, repo domain.TagRepository) []string {
	list, err := repo.List(ctx)
	if err != nil {
		return nil
	}
	return findDuplicates(len(list), func(i int) (string, string, string) {
		return list[i].Name, list[i].Slug, list[i].ID
	}, "标签")
}

// AuditCategoryUniqueness 与 AuditTagUniqueness 对应，针对分类。
func AuditCategoryUniqueness(ctx context.Context, repo domain.CategoryRepository) []string {
	list, err := repo.List(ctx)
	if err != nil {
		return nil
	}
	return findDuplicates(len(list), func(i int) (string, string, string) {
		return list[i].Name, list[i].Slug, list[i].ID
	}, "分类")
}

// findDuplicates 通用的 Name/Slug 重复检测：
// accessor(i) 返回第 i 条记录的 (Name, Slug, ID)；kind 是类型名用于拼接错误消息。
func findDuplicates(n int, accessor func(int) (name, slug, id string), kind string) []string {
	var conflicts []string
	nameSeen := make(map[string]string)
	slugSeen := make(map[string]string)
	for i := range n {
		name, slug, id := accessor(i)
		if name != "" {
			if prevID, ok := nameSeen[name]; ok {
				conflicts = append(conflicts, fmt.Sprintf("%s名 %q 在 ID %s 与 %s 间重复", kind, name, prevID, id))
			} else {
				nameSeen[name] = id
			}
		}
		if slug != "" {
			if prevID, ok := slugSeen[slug]; ok {
				conflicts = append(conflicts, fmt.Sprintf("%s Slug %q 在 ID %s 与 %s 间重复", kind, slug, prevID, id))
			} else {
				slugSeen[slug] = id
			}
		}
	}
	return conflicts
}
