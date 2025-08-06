import { useI18n, type ComposerTranslation } from "vue-i18n";

export const appLang = {
  _t: null as ComposerTranslation | null,
  _l: function (): ComposerTranslation {
    if (!this._t) {
      const { t } = useI18n();
      this._t = t;
    }
    return this._t;
  },
  lang: function (text: string, ...args: unknown[]) {
    return this._l()(text, args);
  },
  sort: function (sort: string): string {
    return this._l()(`sort.${sort}`)
  },
  country: function (country: string | undefined | null): string {
    if (country === '' || country === undefined || country === null) {
      return '';
    }
    return this._l()(`country.${country}`)
  },
  definition: function (definition: string): string {
    return this._l()(`definition.${definition}`)
  },
  stars: function (stars: string): string {
    if (stars == '0') {
      return this._l()(`notStar`)
    }
    return stars + this._l()(`stars`)
  },
  year: function (y: string): string {
    if (y == 'before_2000') {
      return this._l()(`before_2000`)
    } else {
      return y + this._l()(`year`)
    }
  },
  attributeTags: function (attrTag: string): string {
    return this._l()(`attributeTags.${attrTag}`)
  },
}


