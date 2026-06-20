import { type ReactNode } from 'react'
import { Link, useLocation } from '@tanstack/react-router'
import { ChevronRight } from 'lucide-react'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuBadge,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  useSidebar,
} from '@/components/ui/sidebar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu'
import {
  type NavCollapsible,
  type NavItem,
  type NavLink,
  type NavGroup as NavGroupProps,
} from './types'

function navTestKey(value?: string) {
  return value?.trim().toLowerCase().replace(/\s+/g, '-') || 'nav-item'
}

export function NavGroup({ title, testIdKey, items }: NavGroupProps) {
  const { state, isMobile } = useSidebar()
  const href = useLocation({ select: (location) => location.href })
  const groupKey = navTestKey(testIdKey || title)
  return (
    <SidebarGroup data-testid={`sidebar-nav-group-${groupKey}`}>
      <SidebarGroupLabel>{title}</SidebarGroupLabel>
      <SidebarMenu>
        {items.map((item) => {
          const key = `${item.title}-${item.url}`

          if (!item.items)
            return <SidebarMenuLink key={key} item={item} href={href} />

          if (state === 'collapsed' && !isMobile)
            return (
              <SidebarMenuCollapsedDropdown key={key} item={item} href={href} />
            )

          return <SidebarMenuCollapsible key={key} item={item} href={href} />
        })}
      </SidebarMenu>
    </SidebarGroup>
  )
}

function NavBadge({
  children,
  itemKey,
}: {
  children: ReactNode
  itemKey: string
}) {
  return (
    <SidebarMenuBadge
      data-testid={`sidebar-nav-badge-${itemKey}`}
      className='end-2 rounded-full bg-sidebar-accent px-1.5 text-[11px] text-sidebar-accent-foreground'
    >
      {children}
    </SidebarMenuBadge>
  )
}

function SidebarMenuLink({ item, href }: { item: NavLink; href: string }) {
  const { state, isMobile, setOpenMobile } = useSidebar()
  const itemKey = navTestKey(item.testIdKey || item.title)
  const isIconOnly = state === 'collapsed' && !isMobile
  return (
    <SidebarMenuItem>
      <SidebarMenuButton
        data-testid={`sidebar-nav-link-${itemKey}`}
        asChild
        isActive={checkIsActive(href, item)}
        tooltip={item.title}
        size='sm'
        className={
          isIconOnly
            ? item.badge
              ? 'justify-center pe-9'
              : 'justify-center'
            : item.badge
              ? 'pe-9'
              : undefined
        }
      >
        <Link
          to={item.url}
          aria-label={item.title}
          title={item.title}
          onClick={() => setOpenMobile(false)}
        >
          {item.icon && <item.icon />}
          {!isIconOnly ? (
            <span data-testid={`sidebar-nav-label-${itemKey}`}>
              {item.title}
            </span>
          ) : null}
          {item.badge && <NavBadge itemKey={itemKey}>{item.badge}</NavBadge>}
        </Link>
      </SidebarMenuButton>
    </SidebarMenuItem>
  )
}

function SidebarMenuCollapsible({
  item,
  href,
}: {
  item: NavCollapsible
  href: string
}) {
  const { state, isMobile, setOpenMobile } = useSidebar()
  const itemKey = navTestKey(item.testIdKey || item.title)
  const isIconOnly = state === 'collapsed' && !isMobile
  return (
    <Collapsible
      asChild
      defaultOpen={checkIsActive(href, item, true)}
      className='group/collapsible'
    >
      <SidebarMenuItem>
        <CollapsibleTrigger asChild>
          <SidebarMenuButton
            aria-label={item.title}
            tooltip={item.title}
            size='sm'
            className={item.badge ? 'pe-9' : undefined}
          >
            {item.icon && <item.icon />}
            <span data-testid={`sidebar-nav-label-${itemKey}`}>
              {item.title}
            </span>
            {item.badge && <NavBadge itemKey={itemKey}>{item.badge}</NavBadge>}
            <ChevronRight className='ms-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90 rtl:rotate-180' />
          </SidebarMenuButton>
        </CollapsibleTrigger>
        <CollapsibleContent className='CollapsibleContent'>
          <SidebarMenuSub>
            {item.items.map((subItem) => {
              const subItemKey = navTestKey(subItem.testIdKey || subItem.title)
              return (
                <SidebarMenuSubItem key={subItem.title}>
                  <SidebarMenuSubButton
                    asChild
                    isActive={checkIsActive(href, subItem)}
                    className={isIconOnly ? 'justify-center' : undefined}
                  >
                    <Link
                      to={subItem.url}
                      aria-label={subItem.title}
                      title={subItem.title}
                      onClick={() => setOpenMobile(false)}
                    >
                      {subItem.icon && <subItem.icon />}
                      {!isIconOnly ? (
                        <span data-testid={`sidebar-nav-label-${subItemKey}`}>
                          {subItem.title}
                        </span>
                      ) : null}
                      {subItem.badge && (
                        <NavBadge itemKey={subItemKey}>
                          {subItem.badge}
                        </NavBadge>
                      )}
                    </Link>
                  </SidebarMenuSubButton>
                </SidebarMenuSubItem>
              )
            })}
          </SidebarMenuSub>
        </CollapsibleContent>
      </SidebarMenuItem>
    </Collapsible>
  )
}

function SidebarMenuCollapsedDropdown({
  item,
  href,
}: {
  item: NavCollapsible
  href: string
}) {
  const itemKey = navTestKey(item.testIdKey || item.title)
  return (
    <SidebarMenuItem>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <SidebarMenuButton
            aria-label={item.title}
            tooltip={item.title}
            isActive={checkIsActive(href, item)}
            size='sm'
            className={item.badge ? 'justify-center pe-9' : 'justify-center'}
          >
            {item.icon && <item.icon />}
            {item.badge && <NavBadge itemKey={itemKey}>{item.badge}</NavBadge>}
            <ChevronRight className='ms-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90' />
          </SidebarMenuButton>
        </DropdownMenuTrigger>
        <DropdownMenuContent side='right' align='start' sideOffset={4}>
          <DropdownMenuLabel>
            {item.title} {item.badge ? `(${item.badge})` : ''}
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          {item.items.map((sub) => (
            <DropdownMenuItem key={`${sub.title}-${sub.url}`} asChild>
              <Link
                to={sub.url}
                className={`${checkIsActive(href, sub) ? 'bg-secondary' : ''}`}
              >
                {sub.icon && <sub.icon />}
                <span className='max-w-52 text-wrap'>{sub.title}</span>
                {sub.badge && (
                  <span className='ms-auto text-xs'>{sub.badge}</span>
                )}
              </Link>
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    </SidebarMenuItem>
  )
}

function checkIsActive(href: string, item: NavItem, mainNav = false) {
  return (
    href === item.url || // /endpint?search=param
    href.split('?')[0] === item.url || // endpoint
    !!item?.items?.filter((i) => i.url === href).length || // if child nav is active
    (mainNav &&
      href.split('/')[1] !== '' &&
      href.split('/')[1] === item?.url?.split('/')[1])
  )
}
