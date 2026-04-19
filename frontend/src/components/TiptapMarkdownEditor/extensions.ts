import { mergeAttributes, Node, type Extensions } from '@tiptap/core'
import Image from '@tiptap/extension-image'
import { TaskItem, TaskList } from '@tiptap/extension-list'
import { TableKit } from '@tiptap/extension-table'
import { CharacterCount, Placeholder } from '@tiptap/extensions'
import { Markdown } from '@tiptap/markdown'
import StarterKit from '@tiptap/starter-kit'

import {
  blockSentinelPrefix,
  inlineSentinelPrefix,
  matchBlockSentinel,
  matchInlineSentinel,
} from './markdown-passthrough'

const GrideaMore = Node.create({
  name: 'grideaMore',
  group: 'block',
  atom: true,
  selectable: true,
  parseHTML() {
    return [{ tag: 'div[data-gridea-more]' }]
  },
  renderHTML({ HTMLAttributes }) {
    return [
      'div',
      mergeAttributes(HTMLAttributes, { 'data-gridea-more': 'true' }),
      ['span', { class: 'gridea-more__label' }, 'MORE'],
    ]
  },
  markdownTokenName: 'grideaMore',
  parseMarkdown() {
    return { type: 'grideaMore' }
  },
  renderMarkdown() {
    return '<!-- more -->'
  },
  markdownTokenizer: {
    name: 'grideaMore',
    level: 'block',
    start(src) {
      return src.indexOf('<!-- more -->')
    },
    tokenize(src) {
      const match = src.match(/^<!--\s*more\s*-->(?:\n|$)/)
      if (!match) {
        return undefined
      }

      return {
        type: 'grideaMore',
        raw: match[0],
        text: '<!-- more -->',
      }
    },
  },
})

const RawHtmlBlock = Node.create({
  name: 'rawHtmlBlock',
  group: 'block',
  atom: true,
  selectable: true,
  addAttributes() {
    return {
      raw: { default: '' },
      kind: { default: 'html' },
    }
  },
  parseHTML() {
    return [{ tag: 'div[data-raw-html-block]' }]
  },
  renderHTML({ node, HTMLAttributes }) {
    return [
      'div',
      mergeAttributes(HTMLAttributes, {
        'data-raw-html-block': node.attrs.kind,
      }),
      ['div', { class: 'gridea-raw-block__label' }, node.attrs.kind === 'comment' ? 'HTML 注释' : 'HTML 块'],
      ['pre', { class: 'gridea-raw-block__content' }, node.attrs.raw],
    ]
  },
  markdownTokenName: 'html',
  parseMarkdown(token) {
    if (!token.block) {
      return []
    }

    const raw = String(token.raw || token.text || '')
    if (!raw.trim() || raw.trim() === '<!-- more -->') {
      return []
    }

    return {
      type: 'rawHtmlBlock',
      attrs: {
        raw,
        kind: raw.trim().startsWith('<!--') ? 'comment' : 'html',
      },
    }
  },
  renderMarkdown(node) {
    return node.attrs?.raw || ''
  },
})

const RawFootnoteDefinition = Node.create({
  name: 'rawFootnoteDefinition',
  group: 'block',
  atom: true,
  selectable: true,
  addAttributes() {
    return {
      raw: { default: '' },
      kind: { default: 'footnote-definition' },
    }
  },
  parseHTML() {
    return [{ tag: 'div[data-raw-footnote-definition]' }]
  },
  renderHTML({ node, HTMLAttributes }) {
    return [
      'div',
      mergeAttributes(HTMLAttributes, {
        'data-raw-footnote-definition': 'true',
      }),
      ['div', { class: 'gridea-raw-block__label' }, '脚注定义'],
      ['pre', { class: 'gridea-raw-block__content' }, node.attrs.raw],
    ]
  },
  markdownTokenName: 'grideaRawFootnoteDefinition',
  parseMarkdown(token) {
    if (token.kind !== 'footnote-definition') {
      return []
    }

    return {
      type: 'rawFootnoteDefinition',
      attrs: {
        raw: token.originalRaw || '',
        kind: token.kind,
      },
    }
  },
  renderMarkdown(node) {
    return node.attrs?.raw || ''
  },
  markdownTokenizer: {
    name: 'grideaRawFootnoteDefinition',
    level: 'block',
    start(src) {
      return src.indexOf(blockSentinelPrefix)
    },
    tokenize(src) {
      const matched = matchBlockSentinel(src)
      if (!matched) {
        return undefined
      }

      return {
        type: 'grideaRawFootnoteDefinition',
        raw: matched.source,
        text: matched.source,
        originalRaw: matched.raw,
        kind: matched.kind,
      }
    },
  },
})

const RawMarkdownInline = Node.create({
  name: 'rawMarkdownInline',
  group: 'inline',
  inline: true,
  atom: true,
  selectable: true,
  addAttributes() {
    return {
      raw: { default: '' },
      kind: { default: 'html' },
    }
  },
  parseHTML() {
    return [{ tag: 'span[data-raw-markdown-inline]' }]
  },
  renderHTML({ node, HTMLAttributes }) {
    return [
      'span',
      mergeAttributes(HTMLAttributes, {
        'data-raw-markdown-inline': node.attrs.kind,
      }),
      node.attrs.raw,
    ]
  },
  markdownTokenName: 'grideaRawInline',
  parseMarkdown(token) {
    if (token.kind !== 'html' && token.kind !== 'subscript') {
      return []
    }

    return {
      type: 'rawMarkdownInline',
      attrs: {
        raw: token.originalRaw || '',
        kind: token.kind,
      },
    }
  },
  renderMarkdown(node) {
    return node.attrs?.raw || ''
  },
  markdownTokenizer: {
    name: 'grideaRawInline',
    level: 'inline',
    start(src) {
      return src.indexOf(inlineSentinelPrefix)
    },
    tokenize(src) {
      const matched = matchInlineSentinel(src)
      if (!matched) {
        return undefined
      }

      return {
        type: 'grideaRawInline',
        raw: matched.source,
        text: matched.source,
        originalRaw: matched.raw,
        kind: matched.kind,
      }
    },
  },
})

export const createArticleEditorExtensions = (placeholder: string): Extensions => [
  Markdown.configure({
    markedOptions: {
      gfm: true,
    },
  }),
  GrideaMore,
  RawHtmlBlock,
  RawFootnoteDefinition,
  RawMarkdownInline,
  StarterKit.configure({
    link: {
      openOnClick: false,
      autolink: false,
      linkOnPaste: false,
    },
  }),
  TableKit,
  TaskList,
  TaskItem.configure({
    nested: true,
  }),
  Image.configure({
    allowBase64: true,
  }),
  Placeholder.configure({
    placeholder,
  }),
  CharacterCount,
]
