export function browserLocale(): string {
    let lang

    if (navigator.languages && navigator.languages.length) {
        ;[lang] = navigator.languages
    } else {
        lang = navigator.language
    }

    return lang
}
