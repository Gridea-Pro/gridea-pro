const BLOCK_SENTINEL_PREFIX = '\uE000GRIDEA_RAW_BLOCK:'
const INLINE_SENTINEL_PREFIX = '\uE000GRIDEA_RAW_INLINE:'
const SENTINEL_SUFFIX = '\uE001'

export type RawBlockKind = 'footnote-definition'
export type RawInlineKind = 'html' | 'subscript'

interface SentinelMatch<Kind extends string> {
  kind: Kind
  raw: string
  source: string
}

const isFenceBoundary = (line: string, activeFenceChar: string | null) => {
  const match = line.match(/^\s*(`{3,}|~{3,})/)
  if (!match) {
    return null
  }

  const fence = match[1]
  const fenceChar = fence[0]

  if (!activeFenceChar) {
    return fenceChar
  }

  if (activeFenceChar === fenceChar) {
    return ''
  }

  return null
}

const encodeRaw = (raw: string) => encodeURIComponent(raw)
const decodeRaw = (raw: string) => decodeURIComponent(raw)

const createBlockSentinel = (kind: RawBlockKind, raw: string) =>
  `${BLOCK_SENTINEL_PREFIX}${kind}:${encodeRaw(raw)}${SENTINEL_SUFFIX}`

const createInlineSentinel = (kind: RawInlineKind, raw: string) =>
  `${INLINE_SENTINEL_PREFIX}${kind}:${encodeRaw(raw)}${SENTINEL_SUFFIX}`

const matchSentinel = <Kind extends string>(
  source: string,
  prefix: string,
): SentinelMatch<Kind> | null => {
  if (!source.startsWith(prefix)) {
    return null
  }

  const endIndex = source.indexOf(SENTINEL_SUFFIX)
  if (endIndex === -1) {
    return null
  }

  const sourceMatch = source.slice(0, endIndex + SENTINEL_SUFFIX.length)
  const body = sourceMatch.slice(prefix.length, -SENTINEL_SUFFIX.length)
  const separatorIndex = body.indexOf(':')
  if (separatorIndex === -1) {
    return null
  }

  return {
    kind: body.slice(0, separatorIndex) as Kind,
    raw: decodeRaw(body.slice(separatorIndex + 1)),
    source: sourceMatch,
  }
}

const replaceInlineMarkdownPassthrough = (segment: string) => {
  const withHtmlTokens = segment.replace(
    /<!--.*?-->|<\/?[A-Za-z][A-Za-z0-9-]*(?:\s[^<>]*?)?>/g,
    (match) => createInlineSentinel('html', match),
  )

  return withHtmlTokens.replace(
    /(^|[^~])~([^~\n]+?)~(?!~)/g,
    (_, prefix: string, content: string) =>
      `${prefix}${createInlineSentinel('subscript', `~${content}~`)}`,
  )
}

const replaceInlineSegments = (line: string) => {
  let output = ''
  let index = 0

  while (index < line.length) {
    if (line[index] === '`') {
      const backtickStart = index
      while (line[index] === '`') {
        index += 1
      }

      const marker = line.slice(backtickStart, index)
      const closingIndex = line.indexOf(marker, index)

      if (closingIndex === -1) {
        output += line.slice(backtickStart)
        break
      }

      output += line.slice(backtickStart, closingIndex + marker.length)
      index = closingIndex + marker.length
      continue
    }

    let plainEnd = index
    while (plainEnd < line.length && line[plainEnd] !== '`') {
      plainEnd += 1
    }

    output += replaceInlineMarkdownPassthrough(line.slice(index, plainEnd))
    index = plainEnd
  }

  return output
}

const replaceFootnoteDefinitions = (markdown: string) => {
  const lines = markdown.split('\n')
  const output: string[] = []
  let activeFenceChar: string | null = null

  for (let index = 0; index < lines.length;) {
    const line = lines[index]
    const fenceChange = isFenceBoundary(line, activeFenceChar)

    if (fenceChange !== null) {
      activeFenceChar = fenceChange || null
      output.push(line)
      index += 1
      continue
    }

    if (
      !activeFenceChar
      && /^\s{0,3}\[\^[^\]]+\]:/.test(line)
    ) {
      const blockLines = [line]
      index += 1

      while (index < lines.length) {
        const nextLine = lines[index]

        if (/^( {4}|\t)/.test(nextLine)) {
          blockLines.push(nextLine)
          index += 1
          continue
        }

        if (nextLine === '' && index + 1 < lines.length && /^( {4}|\t)/.test(lines[index + 1])) {
          blockLines.push(nextLine)
          index += 1
          continue
        }

        break
      }

      output.push(createBlockSentinel('footnote-definition', blockLines.join('\n')))
      continue
    }

    output.push(line)
    index += 1
  }

  return output.join('\n')
}

const replaceInlinePassthrough = (markdown: string) => {
  const lines = markdown.split('\n')
  const output: string[] = []
  let activeFenceChar: string | null = null

  for (const line of lines) {
    const fenceChange = isFenceBoundary(line, activeFenceChar)

    if (fenceChange !== null) {
      activeFenceChar = fenceChange || null
      output.push(line)
      continue
    }

    if (activeFenceChar || line.includes(BLOCK_SENTINEL_PREFIX)) {
      output.push(line)
      continue
    }

    output.push(replaceInlineSegments(line))
  }

  return output.join('\n')
}

export const preprocessMarkdownForTiptap = (markdown: string) => {
  const normalized = markdown.replace(/\r\n?/g, '\n')
  const withRawBlocks = replaceFootnoteDefinitions(normalized)

  return replaceInlinePassthrough(withRawBlocks)
}

export const matchBlockSentinel = (source: string) =>
  matchSentinel<RawBlockKind>(source, BLOCK_SENTINEL_PREFIX)

export const matchInlineSentinel = (source: string) =>
  matchSentinel<RawInlineKind>(source, INLINE_SENTINEL_PREFIX)

export const blockSentinelPrefix = BLOCK_SENTINEL_PREFIX
export const inlineSentinelPrefix = INLINE_SENTINEL_PREFIX
