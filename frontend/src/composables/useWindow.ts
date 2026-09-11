import { onBeforeUnmount, onMounted, ref } from 'vue'
import {
  WindowCenter,
  WindowIsMaximised,
  WindowMinimise,
  WindowSetSize,
  WindowToggleMaximise,
} from '../../wailsjs/runtime/runtime'
import * as api from '../api/app'

export interface WindowSize {
  width: number
  height: number
}

/** main.go の Width/Height と合わせる */
export const MAIN_SIZE: WindowSize = { width: 1200 , height: 860 }
/** セットアップ画面用のコンパクトサイズ（main.go の MinWidth/MinHeight 以上にする） */
export const SETUP_SIZE: WindowSize = { width: 720, height: 620 }

/** タイトルバー用のウィンドウ操作 */
export function useWindowControls() {
  const isMaximised = ref(false)

  async function refresh() {
    try {
      isMaximised.value = await WindowIsMaximised()
    } catch {
      /* Wails 外（ブラウザ）で開いた場合 */
    }
  }

  onMounted(() => {
    void refresh()
    window.addEventListener('resize', refresh)
  })
  onBeforeUnmount(() => {
    window.removeEventListener('resize', refresh)
  })

  function minimise() {
    try {
      WindowMinimise()
    } catch {
      /* ignore */
    }
  }
  function toggleMaximise() {
    try {
      WindowToggleMaximise()
    } catch {
      /* ignore */
    }
    void refresh()
  }
  /** ウィンドウを閉じる＝Go 側でウィンドウを隠してバックグラウンドで動作を続ける（復帰・終了はタスクトレイから） */
  async function hide() {
    try {
      await api.HideWindow()
    } catch {
      /* ignore */
    }
  }

  return { isMaximised, minimise, toggleMaximise, hide }
}

/** 最大化中でなければウィンドウサイズを変更して中央に置く */
export async function applyWindowSize(size: WindowSize) {
  try {
    if (await WindowIsMaximised()) return
    WindowSetSize(size.width, size.height)
    WindowCenter()
  } catch (e) {
    console.warn('window resize skipped', e)
  }
}
