export * from '../../wailsjs/go/app/App'
export { service, settings } from '../../wailsjs/go/models'

/** Wails のバインディングは Error ではなく文字列で reject するため正規化する */
export function toError(e: unknown): string {
  if (typeof e === 'string') return e
  if (e instanceof Error) return e.message
  return String(e)
}
