import Editor, { EditorProps, useMonaco } from '@monaco-editor/react'
import { useEffect } from 'react'

export default function MonacoEditor({ value, onChange, theme, ...rest }: EditorProps) {
    const monaco = useMonaco()

    useEffect(() => {
        if (!monaco) {
            return undefined
        }
        if (!theme) {
            return undefined
        }
        if (theme === 'vs-dark' || theme === 'light') {
            return undefined
        }
        import(`monaco-themes/themes/${theme}.json`).then((data) => {
            monaco.editor.defineTheme(theme, data)
            monaco.editor.setTheme(theme)
        })
        return undefined
    }, [monaco, theme])

    // eslint-disable-next-line react/jsx-props-no-spreading
    return <Editor theme={theme} value={value} onChange={onChange} {...rest} />
}
