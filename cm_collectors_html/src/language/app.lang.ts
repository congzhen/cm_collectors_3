import { appStoreData } from "@/storeData/app.storeData";
import { t } from "./app.lang.main";
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function AppLang() {
  const storeAppStoreData = appStoreData()
  const translations = {
    sort: (sort: string): string => t(`sort.${sort}`),
    country: (country: string | undefined | null): string => {
      if (country === '' || country === undefined || country === null) {
        return '';
      }

      return t(`country.${country}`)
    },
    definition: (definition: string): string => t(`definition.${definition}`),
    stars: (stars: string): string => {
      if (stars == '0') {
        return t(`notStar`)
      }
      return stars + t(`stars`)
    },

    year: (y: string): string => {
      if (y == 'before_2000') {
        return t(`before_2000`)
      } else {
        return y + t(`year`)
      }
    },

    attributeTags: (attrTag: string): string => t(`attributeTags.${attrTag}`),

    performer: (): string => {
      const performerText = storeAppStoreData.currentConfigApp.performer_Text;
      return performerText == '' ? t(`performer`) : performerText;
    },

    director: (): string => {
      const directorText = storeAppStoreData.currentConfigApp.director_Text;
      return directorText == '' ? t(`director`) : directorText;
    },
    lang: t,
  }
  return {
    ...translations,
    t // 暴露原始的 t 函数以备不时之需
  }
}


