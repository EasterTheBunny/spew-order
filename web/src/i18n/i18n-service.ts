import i18next, { i18n, Resource } from "i18next"
import LanguageDetector from "i18next-browser-languagedetector"
import translations from "./translations"

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
      detection: { order: ['querystring', 'navigator'] },
      debug: false,
      defaultNS: "common",
      fallbackNS: "common",
      fallbackLng: "en",
      supportedLngs: ['en', 'es'],
      interpolation: {
        escapeValue: false,
      },
      resources: translations as Resource,
    })
  }

  public changeLanguage(language: string): void {
    this.i18n.changeLanguage(language)
  }
}
