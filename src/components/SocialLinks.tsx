import type { SiteSetting } from '@/payload-types'

type SocialLink = NonNullable<SiteSetting['socialLinks']>[number]

const platformLabels: Record<SocialLink['platform'], string> = {
  facebook: 'Facebook',
  instagram: 'Instagram',
  other: 'Website',
  tiktok: 'TikTok',
  x: 'X',
  youtube: 'YouTube',
}

type SocialLinksProps = {
  className?: string
  linkClassName?: string
  links: SocialLink[]
}

export function SocialLinks({
  className = '',
  linkClassName = '',
  links,
}: SocialLinksProps) {
  if (links.length === 0) {
    return null
  }

  return (
    <ul className={className}>
      {links.map((link, index) => (
        <li key={link.id ?? `${link.platform}-${index}`}>
          <a
            className={linkClassName}
            href={link.url}
            rel="noreferrer"
            target="_blank"
          >
            {platformLabels[link.platform] ?? 'Link'}
          </a>
        </li>
      ))}
    </ul>
  )
}
