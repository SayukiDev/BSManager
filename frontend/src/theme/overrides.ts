import type { GlobalThemeOverrides } from 'naive-ui'

/** 参考UIのピンクをプライマリカラーにする。他は naive-ui の標準テーマをそのまま使う */
export const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#E91E63',
    primaryColorHover: '#F06292',
    primaryColorPressed: '#C2185B',
    primaryColorSuppl: '#F06292',
    borderRadius: '8px',
  },
}
