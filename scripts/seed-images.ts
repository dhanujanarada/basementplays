import fs from 'node:fs'
import path from 'node:path'

import sharp from 'sharp'

export type ImageKind = 'poster' | 'hero' | 'gallery' | 'backstage'

export type ImageMotif = 'crown' | 'sun' | 'stage' | 'candid'

export type PlaceholderImage = {
  accent: string
  alt: string
  background: string
  eyebrow: string
  foreground: string
  key: string
  kind: ImageKind
  motif: ImageMotif
  subtitle?: string
  title: string
}

const outputDir = path.join('/tmp/opencode', 'basementplays-seed')

const dimensions: Record<ImageKind, { height: number; width: number }> = {
  poster: { height: 1800, width: 1200 },
  hero: { height: 1080, width: 1920 },
  gallery: { height: 1000, width: 1600 },
  backstage: { height: 900, width: 1400 },
}

export async function renderPlaceholder(
  image: PlaceholderImage,
): Promise<string> {
  fs.mkdirSync(outputDir, { recursive: true })

  const { height, width } = dimensions[image.kind]
  const filePath = path.join(outputDir, `${image.key}.jpg`)
  const svg = buildSvg(image, width, height)

  await sharp(Buffer.from(svg))
    .jpeg({ mozjpeg: true, quality: 90 })
    .toFile(filePath)

  return filePath
}

function buildSvg(
  image: PlaceholderImage,
  width: number,
  height: number,
): string {
  const centered = image.kind === 'poster'
  const titleFont =
    image.kind === 'poster' || image.kind === 'hero'
      ? "Georgia, 'Times New Roman', serif"
      : "Inter, Helvetica, Arial, sans-serif"
  const titleSize =
    image.kind === 'poster'
      ? 112
      : image.kind === 'hero'
        ? 94
        : image.kind === 'gallery'
          ? 64
          : 52
  const titleLines = image.title.split('|')
  const lineHeight = titleSize * 1.06
  const titleX = centered ? width / 2 : 96
  const titleAnchor = centered ? 'middle' : 'start'
  const titleStartY =
    image.kind === 'poster'
      ? 840
      : image.kind === 'hero'
        ? 570
        : height - 250
  const subtitleY = titleStartY + titleLines.length * lineHeight + 46
  const subtitleSize = image.kind === 'poster' ? 30 : 24
  const frame = image.kind === 'poster' ? 48 : 56
  const eyebrowY = image.kind === 'poster' ? 156 : 120
  const eyebrowX = centered ? width / 2 : 96
  const eyebrowAnchor = centered ? 'middle' : 'start'

  const titleText = titleLines
    .map(
      (line, index) =>
        `<tspan x="${titleX}" y="${titleStartY + index * lineHeight}">${escapeXml(line)}</tspan>`,
    )
    .join('')

  return `
    <svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}">
      <defs>
        <radialGradient id="glow" cx="50%" cy="32%" r="58%">
          <stop offset="0%" stop-color="${image.accent}" stop-opacity="0.32" />
          <stop offset="62%" stop-color="${image.accent}" stop-opacity="0.08" />
          <stop offset="100%" stop-color="${image.background}" stop-opacity="0" />
        </radialGradient>
        <pattern id="dots" width="28" height="28" patternUnits="userSpaceOnUse">
          <circle cx="3" cy="3" r="1.1" fill="${image.foreground}" opacity="0.10" />
        </pattern>
      </defs>

      <rect width="${width}" height="${height}" fill="${image.background}" />
      <rect width="${width}" height="${height}" fill="url(#glow)" />
      <rect width="${width}" height="${height}" fill="url(#dots)" />
      <rect width="${width}" height="${height}" fill="#000000" opacity="0.08" />

      ${motifSvg(image.motif, width, height, image.accent, image.foreground)}

      <rect
        x="${frame}"
        y="${frame}"
        width="${width - frame * 2}"
        height="${height - frame * 2}"
        fill="none"
        stroke="${image.accent}"
        stroke-opacity="0.42"
        stroke-width="2"
      />

      <text
        x="${eyebrowX}"
        y="${eyebrowY}"
        text-anchor="${eyebrowAnchor}"
        fill="${image.accent}"
        font-family="Inter, Helvetica, Arial, sans-serif"
        font-size="${image.kind === 'poster' ? 24 : 20}"
        font-weight="600"
        letter-spacing="7"
      >${escapeXml(image.eyebrow.toUpperCase())}</text>

      <text
        text-anchor="${titleAnchor}"
        fill="${image.foreground}"
        font-family="${titleFont}"
        font-size="${titleSize}"
        font-weight="700"
        letter-spacing="${image.kind === 'poster' ? -3 : -1}"
        paint-order="stroke"
        stroke="${image.background}"
        stroke-width="10"
        stroke-linejoin="round"
      >${titleText}</text>

      ${
        image.subtitle
          ? `<text
              x="${titleX}"
              y="${subtitleY}"
              text-anchor="${titleAnchor}"
              fill="${image.foreground}"
              fill-opacity="0.76"
              font-family="Inter, Helvetica, Arial, sans-serif"
              font-size="${subtitleSize}"
              font-weight="400"
              letter-spacing="2"
            >${escapeXml(image.subtitle)}</text>`
          : ''
      }

      <line
        x1="${frame}"
        y1="${height - frame - 54}"
        x2="${width - frame}"
        y2="${height - frame - 54}"
        stroke="${image.foreground}"
        stroke-opacity="0.18"
        stroke-width="1"
      />

      <text
        x="${frame}"
        y="${height - frame - 18}"
        fill="${image.foreground}"
        fill-opacity="0.62"
        font-family="Inter, Helvetica, Arial, sans-serif"
        font-size="18"
        font-weight="500"
        letter-spacing="5"
      >BASEMENT PLAYS</text>

      <text
        x="${width - frame}"
        y="${height - frame - 18}"
        text-anchor="end"
        fill="${image.accent}"
        fill-opacity="0.86"
        font-family="Inter, Helvetica, Arial, sans-serif"
        font-size="18"
        font-weight="600"
        letter-spacing="4"
      >2026 SEASON</text>
    </svg>
  `
}

function motifSvg(
  motif: ImageMotif,
  width: number,
  height: number,
  accent: string,
  foreground: string,
): string {
  switch (motif) {
    case 'crown':
      return `
        <g
          transform="translate(${width * 0.5} ${height * 0.28})"
          fill="none"
          stroke="${accent}"
          stroke-linejoin="round"
          stroke-width="8"
          opacity="0.62"
        >
          <path d="M -190 90 L -150 -70 L -62 18 L 0 -118 L 62 18 L 150 -70 L 190 90 Z" />
          <path d="M -190 90 L 190 90" />
          <circle cx="-150" cy="-70" r="14" fill="${accent}" />
          <circle cx="0" cy="-118" r="14" fill="${accent}" />
          <circle cx="150" cy="-70" r="14" fill="${accent}" />
        </g>
      `
    case 'sun':
      return `
        <g
          transform="translate(${width * 0.5} ${height * 0.28})"
          stroke="${accent}"
          stroke-linecap="round"
          stroke-width="10"
          opacity="0.58"
        >
          <circle r="142" fill="${accent}" fill-opacity="0.16" />
          ${Array.from({ length: 16 }, (_, index) => {
            const angle = (index / 16) * Math.PI * 2
            const inner = 196
            const outer = 286

            return `<line x1="${Math.cos(angle) * inner}" y1="${Math.sin(angle) * inner}" x2="${Math.cos(angle) * outer}" y2="${Math.sin(angle) * outer}" />`
          }).join('')}
        </g>
      `
    case 'stage':
      return `
        <g opacity="0.42">
          <path
            d="M ${width * 0.16} 0 L ${width * 0.4} ${height * 0.74} L ${width * 0.6} ${height * 0.74} L ${width * 0.84} 0 Z"
            fill="${accent}"
            opacity="0.12"
          />
          <path
            d="M ${width * 0.34} 0 L ${width * 0.46} ${height * 0.74} L ${width * 0.54} ${height * 0.74} L ${width * 0.66} 0 Z"
            fill="${foreground}"
            opacity="0.06"
          />
          <line
            x1="${width * 0.08}"
            y1="${height * 0.8}"
            x2="${width * 0.92}"
            y2="${height * 0.8}"
            stroke="${accent}"
            stroke-width="4"
          />
        </g>
      `
    case 'candid':
      return `
        <g opacity="0.34" stroke="${foreground}" fill="none">
          <rect
            x="${width * 0.09}"
            y="${height * 0.16}"
            width="${width * 0.82}"
            height="${height * 0.68}"
            rx="18"
            stroke-dasharray="14 18"
            stroke-width="3"
          />
          <circle
            cx="${width * 0.5}"
            cy="${height * 0.5}"
            r="${height * 0.16}"
            stroke="${accent}"
            stroke-width="4"
          />
          <line
            x1="${width * 0.5}"
            y1="${height * 0.16}"
            x2="${width * 0.5}"
            y2="${height * 0.84}"
            stroke="${foreground}"
            stroke-width="2"
          />
          <line
            x1="${width * 0.09}"
            y1="${height * 0.5}"
            x2="${width * 0.91}"
            y2="${height * 0.5}"
            stroke="${foreground}"
            stroke-width="2"
          />
        </g>
      `
  }
}

function escapeXml(value: string): string {
  return value.replace(/[<>&'"]/g, (character) => {
    switch (character) {
      case '<':
        return '&lt;'
      case '>':
        return '&gt;'
      case '&':
        return '&amp;'
      case "'":
        return '&apos;'
      case '"':
        return '&quot;'
      default:
        return character
    }
  })
}
