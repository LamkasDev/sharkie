package ime

import . "github.com/LamkasDev/sharkie/cmd/lib_structs/user"

type ImeType uint32

const (
	ImeTypeDefault    = ImeType(0)
	ImeTypeBasicLatin = ImeType(1)
	ImeTypeUrl        = ImeType(2)
	ImeTypeMail       = ImeType(3)
	ImeTypeNumber     = ImeType(4)
)

type ImeLanguage uint64

const (
	ImeLanguageDanish             = ImeLanguage(1 << 0)
	ImeLanguageGerman             = ImeLanguage(1 << 1)
	ImeLanguageEnglishUS          = ImeLanguage(1 << 2)
	ImeLanguageSpanish            = ImeLanguage(1 << 3)
	ImeLanguageFrench             = ImeLanguage(1 << 4)
	ImeLanguageItalian            = ImeLanguage(1 << 5)
	ImeLanguageDutch              = ImeLanguage(1 << 6)
	ImeLanguageNorwegian          = ImeLanguage(1 << 7)
	ImeLanguagePolish             = ImeLanguage(1 << 8)
	ImeLanguagePortuguesePT       = ImeLanguage(1 << 9)
	ImeLanguageRussian            = ImeLanguage(1 << 10)
	ImeLanguageFinnish            = ImeLanguage(1 << 11)
	ImeLanguageSwedish            = ImeLanguage(1 << 12)
	ImeLanguageJapanese           = ImeLanguage(1 << 13)
	ImeLanguageKorean             = ImeLanguage(1 << 14)
	ImeLanguageSimplifiedChinese  = ImeLanguage(1 << 15)
	ImeLanguageTraditionalChinese = ImeLanguage(1 << 16)
	ImeLanguagePortugueseBR       = ImeLanguage(1 << 17)
	ImeLanguageEnglishGB          = ImeLanguage(1 << 18)
	ImeLanguageTurkish            = ImeLanguage(1 << 19)
	ImeLanguageSpanishLA          = ImeLanguage(1 << 20)
	ImeLanguageArabic             = ImeLanguage(1 << 24)
	ImeLanguageFrenchCA           = ImeLanguage(1 << 25)
	ImeLanguageThai               = ImeLanguage(1 << 26)
	ImeLanguageCzech              = ImeLanguage(1 << 27)
	ImeLanguageGreek              = ImeLanguage(1 << 28)
	ImeLanguageIndonesian         = ImeLanguage(1 << 29)
	ImeLanguageVietnamese         = ImeLanguage(1 << 30)
	ImeLanguageRomanian           = ImeLanguage(1 << 31)
	ImeLanguageHungarian          = ImeLanguage(1 << 32)
)

type ImeEnterLabel uint32

const (
	ImeEnterLabelDefault = ImeEnterLabel(0)
	ImeEnterLabelSend    = ImeEnterLabel(1)
	ImeEnterLabelSearch  = ImeEnterLabel(2)
	ImeEnterLabelGo      = ImeEnterLabel(3)
)

type ImeInputMethod uint32

const (
	ImeInputMethodDefault = ImeInputMethod(0)
)

type ImeOption uint32

const (
	ImeOptionDefault                       = ImeOption(0)
	ImeOptionMultiline                     = ImeOption(1 << 0)
	ImeOptionNoAutoCapitalization          = ImeOption(1 << 1)
	ImeOptionPassword                      = ImeOption(1 << 2)
	ImeOptionLanguagesForced               = ImeOption(1 << 3)
	ImeOptionExtKeyboard                   = ImeOption(1 << 4)
	ImeOptionNoLearning                    = ImeOption(1 << 5)
	ImeOptionFixedPosition                 = ImeOption(1 << 6)
	ImeOptionDisableCopyPaste              = ImeOption(1 << 7)
	ImeOptionDisableResume                 = ImeOption(1 << 8)
	ImeOptionDisableAutoSpace              = ImeOption(1 << 9)
	ImeOptionDisablePositionAdjustment     = ImeOption(1 << 11)
	ImeOptionExpandedPreeditBuffer         = ImeOption(1 << 12)
	ImeOptionUseJapaneseEisuuKeyAsCapslock = ImeOption(1 << 13)
	ImeOptionUseOver2kCoordinates          = ImeOption(1 << 14)
)

type ImeHorizontalAlignment uint32

const (
	ImeHorizontalAlignmentLeft   = ImeHorizontalAlignment(0)
	ImeHorizontalAlignmentCenter = ImeHorizontalAlignment(1)
	ImeHorizontalAlignmentRight  = ImeHorizontalAlignment(2)
)

type ImeVerticalAlignment uint32

const (
	ImeVerticalAlignmentTop    = ImeVerticalAlignment(0)
	ImeVerticalAlignmentCenter = ImeVerticalAlignment(1)
	ImeVerticalAlignmentBottom = ImeVerticalAlignment(2)
)

type ImeParam struct {
	UserId              UserId
	Type                ImeType
	SupportedLanguages  ImeLanguage
	EnterLabel          ImeEnterLabel
	InputMethod         ImeInputMethod
	TextFilterAddress   uintptr
	Options             ImeOption
	MaxTextLength       uint32
	InputTextBuffer     *uint16
	PosX                float32
	PosY                float32
	HorizontalAlignment ImeHorizontalAlignment
	VerticalAlignment   ImeVerticalAlignment
	WorkPtr             uintptr
	ArgPtr              uintptr
	EventHandlerAddress uintptr
	Reserved            [8]byte
}
