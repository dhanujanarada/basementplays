import type { Run } from '@/payload-types'

import type { Show, ShowEntry } from './types'

function startTime(show: Show): number {
  return new Date(show.startsAt).getTime()
}

function latestShowTime(run: Run): number {
  const times = (run.shows ?? [])
    .map(startTime)
    .filter((time) => !Number.isNaN(time))

  return times.length > 0 ? Math.max(...times) : 0
}

export function getUpcomingShow(
  runs: Run[],
  now: Date = new Date(),
): ShowEntry | null {
  const nowTime = now.getTime()
  const upcoming: ShowEntry[] = []

  for (const run of runs) {
    for (const show of run.shows ?? []) {
      const time = startTime(show)

      if (
        show.status === 'cancelled' ||
        Number.isNaN(time) ||
        time < nowTime
      ) {
        continue
      }

      upcoming.push({ run, show })
    }
  }

  upcoming.sort((a, b) => startTime(a.show) - startTime(b.show))

  return upcoming[0] ?? null
}

export function sortShowsAscending(shows: Show[]): Show[] {
  return [...shows].sort((a, b) => startTime(a) - startTime(b))
}

export function sortRunsNewestFirst(runs: Run[]): Run[] {
  return [...runs].sort((a, b) => latestShowTime(b) - latestShowTime(a))
}
