import LabelList from './LabelList'

export interface IMapInputProps {
    value?: Record<string, string>
    onChange?: (value: Record<string, string>) => void
    style?: React.CSSProperties
}

export default function MapInput(props: IMapInputProps) {
    const { value = {}, onChange, style } = props

    return (
        <LabelList
            style={style}
            value={Object.keys(value).map((key) => ({ key, value: value[key] }))}
            onChange={async (v) => {
                onChange?.(
                    v.reduce((acc: Record<string, string>, cur) => {
                        return {
                            ...acc,
                            [cur.key]: cur.value,
                        }
                    }, {})
                )
            }}
        />
    )
}
