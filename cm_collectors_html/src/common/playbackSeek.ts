// Keep a single pending target. The first click is immediate; a burst of clicks
// or range input events can interrupt decoding at most once per interval.
export function createPlaybackSeek(apply: (time: number) => void, interval = 120) {
  let timer: ReturnType<typeof setTimeout> | undefined
  let target: number | undefined
  let lastApplied = -Infinity
  const commit = () => {
    timer = undefined
    if (target === undefined) return
    const time = target
    target = undefined
    lastApplied = Date.now()
    apply(time)
  }
  return {
    request(time: number) {
      if (!Number.isFinite(time)) return
      target = time
      const delay = interval - (Date.now() - lastApplied)
      if (delay <= 0) {
        clearTimeout(timer)
        commit()
      } else if (timer === undefined) {
        timer = setTimeout(commit, delay)
      }
    },
    pending: () => target !== undefined,
    cancel() {
      clearTimeout(timer)
      timer = undefined
      target = undefined
      lastApplied = -Infinity
    },
  }
}
