package serial

import (
	"strings"

	"github.com/go-playground/locales/ar"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/es"
	"github.com/go-playground/locales/fa"
	"github.com/go-playground/locales/fr"
	"github.com/go-playground/locales/id"
	"github.com/go-playground/locales/it"
	"github.com/go-playground/locales/ja"
	"github.com/go-playground/locales/nl"
	"github.com/go-playground/locales/pt"
	"github.com/go-playground/locales/pt_BR"
	"github.com/go-playground/locales/ru"
	"github.com/go-playground/locales/tr"
	"github.com/go-playground/locales/vi"
	"github.com/go-playground/locales/zh"
	"github.com/go-playground/locales/zh_Hant_TW"
	translator "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ar_translations "github.com/go-playground/validator/v10/translations/ar"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	es_translations "github.com/go-playground/validator/v10/translations/es"
	fa_translations "github.com/go-playground/validator/v10/translations/fa"
	fr_translations "github.com/go-playground/validator/v10/translations/fr"
	id_translations "github.com/go-playground/validator/v10/translations/id"
	it_translations "github.com/go-playground/validator/v10/translations/it"
	ja_translations "github.com/go-playground/validator/v10/translations/ja"
	nl_translations "github.com/go-playground/validator/v10/translations/nl"
	pt_translations "github.com/go-playground/validator/v10/translations/pt"
	pt_BR_translations "github.com/go-playground/validator/v10/translations/pt_BR"
	ru_translations "github.com/go-playground/validator/v10/translations/ru"
	tr_translations "github.com/go-playground/validator/v10/translations/tr"
	vi_translations "github.com/go-playground/validator/v10/translations/vi"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	zh_tw_translations "github.com/go-playground/validator/v10/translations/zh_tw"
)

var languages = make(map[string]string)

func init() {
	languages["ar"] = "ar"
	languages["en"] = "en"
	languages["es"] = "es"
	languages["fa"] = "fa"
	languages["fr"] = "fr"
	languages["id"] = "id"
	languages["it"] = "it"
	languages["ja"] = "ja"
	languages["nl"] = "nl"
	languages["pt"] = "pt"
	languages["pt_br"] = "pt_BR"
	languages["ru"] = "ru"
	languages["tr"] = "tr"
	languages["vi"] = "vi"
	languages["zh"] = "zh"
	languages["zh_tw"] = "zh_Hant_TW"
}

func universalTranslator(validate *validator.Validate, language string) translator.Translator {
	ar := ar.New()
	en := en.New()
	es := es.New()
	fa := fa.New()
	fr := fr.New()
	id := id.New()
	it := it.New()
	ja := ja.New()
	nl := nl.New()
	pt := pt.New()
	pt_BR := pt_BR.New()
	ru := ru.New()
	tr := tr.New()
	vi := vi.New()
	zh := zh.New()
	zh_Hant_TW := zh_Hant_TW.New()
	universal := translator.New(ar, en, es, fa, fr, id, it, ja, nl, pt, pt_BR, ru, tr, vi, zh, zh_Hant_TW)
	if lang, exist := languages[strings.ToLower(language)]; exist {
		trans, ok := universal.GetTranslator(lang)
		if !ok {
			return nil
		}
		switch {
		case strings.EqualFold(lang, "ar"):
			ar_translations.RegisterDefaultTranslations(validate, trans)
		case strings.EqualFold(lang, "en"):
			en_translations.RegisterDefaultTranslations(validate, trans)
		case strings.EqualFold(lang, "es"):
			es_translations.RegisterDefaultTranslations(validate, trans)
		case strings.EqualFold(lang, "fa"):
			fa_translations.RegisterDefaultTranslations(validate, trans)
		case strings.EqualFold(lang, "fr"):
			fr_translations.RegisterDefaultTranslations(validate, trans)
		case strings.EqualFold(lang, "id"):
			id_translations.RegisterDefaultTranslations(validate, trans)
		case strings.EqualFold(lang, "it"):
			it_translations.RegisterDefaultTranslations(validate, trans)
		case strings.EqualFold(lang, "ja"):
			ja_translations.RegisterDefaultTranslations(validate, trans)
		case strings.EqualFold(lang, "nl"):
			nl_translations.RegisterDefaultTranslations(validate, trans)
		case strings.EqualFold(lang, "pt"):
			pt_translations.RegisterDefaultTranslations(validate, trans)
		case strings.EqualFold(lang, "pt_br"):
			pt_BR_translations.RegisterDefaultTranslations(validate, trans)
		case strings.EqualFold(lang, "ru"):
			ru_translations.RegisterDefaultTranslations(validate, trans)
		case strings.EqualFold(lang, "tr"):
			tr_translations.RegisterDefaultTranslations(validate, trans)
		case strings.EqualFold(lang, "vi"):
			vi_translations.RegisterDefaultTranslations(validate, trans)
		case strings.EqualFold(lang, "zh"):
			zh_translations.RegisterDefaultTranslations(validate, trans)
		case strings.EqualFold(lang, "zh_tw"):
			zh_tw_translations.RegisterDefaultTranslations(validate, trans)
		}
		return trans
	}
	return nil
}
