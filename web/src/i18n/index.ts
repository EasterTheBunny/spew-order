import { getContext, setContext } from "svelte"
import { I18nService } from "./i18n-service"
import { I18NextTranslationService } from "./translation-service"

const CONTEXT_KEY = "t"

export const initLocalizationContext: () => I18nService = () => {
  // Initialize our services
  const i18n = new I18nService()
  const translator = new I18NextTranslationService(i18n)

  // Setting the Svelte context
  setLocalization({
    currentLanguage: translator.locale,
    t: translator.translate,
  })

  return i18n
}

export const setLocalization: (context: I18nContext) => void = (context: I18nContext) => {
  return setContext<I18nContext>(CONTEXT_KEY, context)
}

// To make retrieving the t function easier.
export const getLocalization: () => I18nContext = () => {
  return getContext<I18nContext>(CONTEXT_KEY)
}
