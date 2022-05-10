import { docco, dark } from 'react-syntax-highlighter/dist/esm/styles/hljs'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import SyntaxHighlighter from 'react-syntax-highlighter'

export interface ICopyableTextProps {
    children: React.ReactNode
}

export default function HighlightText({ children }: ICopyableTextProps) {
    const themeType = useCurrentThemeType()
    const highlightTheme = themeType === 'dark' ? dark : docco

    return (
        <SyntaxHighlighter
            language='bash'
            style={highlightTheme}
            customStyle={{
                display: 'inline',
                margin: 0,
                padding: '0px 2px',
                borderRadius: '2px',
            }}
        >
            {children}
        </SyntaxHighlighter>
    )
}
