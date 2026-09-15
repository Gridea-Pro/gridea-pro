import MarkdownIt from 'markdown-it'
import MarkdownItKatex from '@iktakahiro/markdown-it-katex'
import markdownItTocAndAnchor from 'markdown-it-toc-and-anchor'
import MarkdownItTaskLists from 'markdown-it-task-lists'
import MarkdownItMark from 'markdown-it-mark'
import MarkdownItSup from 'markdown-it-sup'
import MarkdownItSub from 'markdown-it-sub'
import MarkdownItAbbr from 'markdown-it-abbr'
import MarkdownItFootnote from 'markdown-it-footnote'
// import MarkdownItImsize from 'markdown-it-imsize' // Disabled for browser compatibility
import { full as MarkdownItEmoji } from 'markdown-it-emoji'
import MarkdownItImplicitFigures from 'markdown-it-implicit-figures'
// import MarkdownItImageLazyLoading from 'markdown-it-image-lazy-loading' // Disabled for browser compatibility as it uses 'fs'
import DOMPurify from 'dompurify'

const markdownIt = new MarkdownIt({
  html: true,
  breaks: true,
})

const BAD_PROTO_RE = /^(vbscript|javascript|data):/
const GOOD_DATA_RE = /^data:image\/(gif|png|jpeg|webp);/

markdownIt.validateLink = function (url) {
  url = url.trim().toLowerCase()

  return BAD_PROTO_RE.test(url) ? (!!GOOD_DATA_RE.test(url)) : true
}

markdownIt.use(MarkdownItKatex)
markdownIt.use(markdownItTocAndAnchor, {
  anchorLink: false,
})
markdownIt.use(MarkdownItTaskLists, {
  label: true,
  labelAfter: true,
})
markdownIt.use(MarkdownItMark)
markdownIt.use(MarkdownItSup)
markdownIt.use(MarkdownItSub)
markdownIt.use(MarkdownItAbbr)
markdownIt.use(MarkdownItFootnote)
// markdownIt.use(MarkdownItImsize)
markdownIt.use(MarkdownItEmoji)
markdownIt.use(MarkdownItImplicitFigures, {
  dataType: true, // <figure data-type="image">, default: false
  figcaption: false, // <figcaption>alternative text</figcaption>, default: false
  tabindex: true, // <figure tabindex="1+n">..., default: false
  link: false, // <a href="img.png"><img src="img.png"></a>, default: false
})
// markdownIt.use(MarkdownItImageLazyLoading)

// 评论专用渲染器：评论内容来自第三方评论平台的访客输入，完全不可信。
// 用 html:false（不解析源码里的原始 HTML 标签）+ DOMPurify 净化最终输出，双重防止存储型 XSS。
// 在 Wails 桌面环境里，评论区一旦 XSS 可调用 window.go.* 后端方法，危害等同本地任意文件读写。
const commentMd = new MarkdownIt({ html: false, breaks: true, linkify: true })
commentMd.validateLink = markdownIt.validateLink

export function renderCommentMarkdown(raw: string): string {
  return DOMPurify.sanitize(commentMd.render(raw))
}

export default markdownIt
