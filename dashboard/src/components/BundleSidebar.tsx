import { useBundle, useBundleLoading } from '@/hooks/useBundle'
import useTranslation from '@/hooks/useTranslation'
import { RiSurveyLine } from 'react-icons/ri'
import React, { useEffect, useMemo } from 'react'
import { useQuery } from 'react-query'
import { useParams } from 'react-router-dom'
import BaseSidebar, { IComposedSidebarProps, INavItem } from '@/components/BaseSidebar'
import { fetchBundle } from '@/services/bundle'
import { useOrganization } from '@/hooks/useOrganization'
import { resourceIconMapping } from '@/consts'

export default function BundleSidebar({ style }: IComposedSidebarProps) {
    const { orgName, bundleName } = useParams<{ orgName: string; bundleName: string }>()
    const bundleInfo = useQuery(`fetchBundle:${orgName}:${bundleName}`, () => fetchBundle(orgName, bundleName))
    const { bundle, setBundle } = useBundle()
    const { organization, setOrganization } = useOrganization()
    const { setBundleLoading } = useBundleLoading()
    useEffect(() => {
        setBundleLoading(bundleInfo.isLoading)
        if (bundleInfo.isSuccess) {
            if (bundleInfo.data.uid !== bundle?.uid) {
                setBundle(bundleInfo.data)
            }
            if (bundleInfo.data.organization?.uid !== organization?.uid) {
                setOrganization(bundleInfo.data.organization)
            }
        } else if (bundleInfo.isLoading) {
            setBundle(undefined)
        }
    }, [
        bundle?.uid,
        bundleInfo.data,
        bundleInfo.isLoading,
        bundleInfo.isSuccess,
        organization?.uid,
        setBundle,
        setBundleLoading,
        setOrganization,
    ])

    const [t] = useTranslation()

    const navItems: INavItem[] = useMemo(
        () => [
            {
                title: t('overview'),
                path: `/orgs/${orgName}/bundles/${bundleName}`,
                icon: RiSurveyLine,
            },
            {
                title: t('sth list', [t('version')]),
                path: `/orgs/${orgName}/bundles/${bundleName}/versions`,
                icon: resourceIconMapping.bundle_version,
            },
        ],
        [bundleName, orgName, t]
    )
    return <BaseSidebar title={bundleName} icon={resourceIconMapping.bundle} navItems={navItems} style={style} />
}
