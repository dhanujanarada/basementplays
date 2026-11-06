export function PlainSectionHeading({
  eyebrow,
  headingId,
  title,
}: {
  eyebrow: string
  headingId: string
  title: string
}) {
  return (
    <div>
      <p className="text-[10px] font-medium uppercase tracking-[0.22em] text-[var(--accent)]">
        {eyebrow}
      </p>
      <h2
        className="font-display mt-3 text-4xl text-[var(--text)] sm:text-5xl"
        id={headingId}
      >
        {title}
      </h2>
    </div>
  )
}
