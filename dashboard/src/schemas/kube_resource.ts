export interface IKubeResourceSchema {
    api_version: string
    kind: string
    name: string
    namespace: string
    match_labels: Record<string, string>
}
