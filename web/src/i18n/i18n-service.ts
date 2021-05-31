import i18next, { i18n, Resource } from "i18next"
import LanguageDetector from "i18next-browser-languagedetector"
import translations from "./translations"

const INITIAL_LANGUAGE = "en"
export class I18nService {
  // expose i18next
  public i18n: i18n

  constructor() {
    this.i18n = i18next.use(LanguageDetector)
    this.initialize()
  }

  // Our translation function
  public t(key: string, replacements?: Record<string, unknown>): string {
    return this.i18n.t(key, replacements)
  }

  // Initializing i18n
  public initialize(): void {
    this.i18n.init({
      debug: false,
      defaultNS: "common",
      fallbackLng: "en",
      fallbackNS: "common",
      interpolation: {
        escapeValue: false,
      },
      lng: INITIAL_LANGUAGE,
      resources: translations as Resource,
    })
  }

  public changeLanguage(language: string): void {
    this.i18n.changeLanguage(language)
  }
}
