import { computed } from 'vue'
import { createDiscreteApi, type MessageApi } from 'naive-ui'
import { useThemeStore } from '../stores/theme'
import { themeOverrides } from '../theme/overrides'

let messageApi: MessageApi | null = null

/**
 * コンポーネントツリー外（Pinia ストアやイベントコールバック）からトーストを出すための API。
 * Pinia 初期化後に呼ばれる必要があるため遅延生成する。
 */
export function getMessage(): MessageApi {
  if (!messageApi) {
    const theme = useThemeStore()
    const { message } = createDiscreteApi(['message'], {
      configProviderProps: computed(() => ({ theme: theme.theme, themeOverrides })),
    })
    messageApi = message
  }
  return messageApi
}
