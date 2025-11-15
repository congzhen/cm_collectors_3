import { appStoreData } from "@/storeData/app.storeData";
import { computed } from "vue";


export const contentLayoutStyle_C = computed(() => {
  const store = {
    appStoreData: appStoreData(),
  }
  return {
    gap: (store.appStoreData.currentConfigApp.coverPosterGap || 4.8) + 'px',
    justifyContent: store.appStoreData.currentConfigApp.resourceJustifyContent || 'flex-start',
  }
});
