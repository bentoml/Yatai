import Editor, { EditorProps, useMonaco, loader } from '@monaco-editor/react'
import { useEffect } from 'react'

loader
    .init()
    .then((monaco) => {
        fetch('/api/v1/deployment_creation_json_schema').then((resp) => {
            resp.json().then((jsn) => {
                monaco.languages.json.jsonDefaults.setDiagnosticsOptions({
                    validate: true,
                    schemas: [
                        {
                            uri: 'http://myserver/foo-schema.json',
                            fileMatch: ['*'],
                            schema: jsn,
                        },
                    ],
                })
            })
        })
    })
    // eslint-disable-next-line no-console
    .catch((error) => console.error('An error occurred during initialization of Monaco: ', error))

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
