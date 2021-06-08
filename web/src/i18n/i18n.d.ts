interface ITranslationService {
  locale: Writable<string>
  translate: Readable<TType>
}

type TType = (text: string, replacements?: Record<string, unknown>) => string

interface I18nContext {
  t: Readable<TType>
  currentLanguage: Writable<string>
}

abstract class I18nService {
  abstract public i18n: i18n
  constructor() {
    // empty constructor
  }
  abstract public t(key: string, replacements?: Record<string, unknown>): string
  abstract public initialize(): void
  abstract public changeLanguage(language: string): void
}
