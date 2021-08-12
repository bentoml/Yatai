import { sidebarExpandedWidth, sidebarFoldedWidth } from '@/consts'
import { SidebarContext } from '@/contexts/SidebarContext'
import { useContext } from 'react'

export default () => {
    const ctx = useContext(SidebarContext)
    return ctx.expanded ? sidebarExpandedWidth : sidebarFoldedWidth
}
