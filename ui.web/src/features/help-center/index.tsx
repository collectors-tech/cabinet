import { useMemo } from 'react'
import { getRouteApi } from '@tanstack/react-router'
import { BookOpenText, FileText, LibraryBig, SearchIcon } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { type HelpCenterArticle, helpCenterArticles } from './articles'

const route = getRouteApi('/_authenticated/help-center/')

type MarkdownBlock =
  | { type: 'heading'; level: 1 | 2 | 3; text: string }
  | { type: 'paragraph'; text: string }
  | { type: 'unordered-list'; items: string[] }
  | { type: 'ordered-list'; items: string[] }

function parseMarkdown(content: string): MarkdownBlock[] {
  const lines = content.split(/\r?\n/)
  const blocks: MarkdownBlock[] = []
  let paragraphLines: string[] = []
  let listItems: string[] = []
  let listType: 'unordered-list' | 'ordered-list' | null = null

  const flushParagraph = () => {
    if (paragraphLines.length === 0) {
      return
    }

    blocks.push({
      type: 'paragraph',
      text: paragraphLines.join(' ').trim(),
    })
    paragraphLines = []
  }

  const flushList = () => {
    if (!listType || listItems.length === 0) {
      listItems = []
      listType = null
      return
    }

    blocks.push({
      type: listType,
      items: [...listItems],
    })

    listItems = []
    listType = null
  }

  for (const rawLine of lines) {
    const line = rawLine.trim()

    if (!line) {
      flushParagraph()
      flushList()
      continue
    }

    const headingMatch = line.match(/^(#{1,3})\s+(.*)$/)
    if (headingMatch) {
      flushParagraph()
      flushList()
      blocks.push({
        type: 'heading',
        level: headingMatch[1].length as 1 | 2 | 3,
        text: headingMatch[2].trim(),
      })
      continue
    }

    const unorderedMatch = line.match(/^[-*]\s+(.*)$/)
    if (unorderedMatch) {
      flushParagraph()
      if (listType && listType !== 'unordered-list') {
        flushList()
      }
      listType = 'unordered-list'
      listItems.push(unorderedMatch[1].trim())
      continue
    }

    const orderedMatch = line.match(/^\d+\.\s+(.*)$/)
    if (orderedMatch) {
      flushParagraph()
      if (listType && listType !== 'ordered-list') {
        flushList()
      }
      listType = 'ordered-list'
      listItems.push(orderedMatch[1].trim())
      continue
    }

    if (listType) {
      flushList()
    }
    paragraphLines.push(line)
  }

  flushParagraph()
  flushList()

  return blocks
}

function MarkdownArticle({ article }: { article: HelpCenterArticle }) {
  const blocks = useMemo(
    () => parseMarkdown(article.content),
    [article.content]
  )

  return (
    <div
      className='space-y-4'
      data-testid={`help-center-article-content-${article.id}`}
    >
      {blocks.map((block, index) => {
        const key = `${article.id}-${block.type}-${index}`

        if (block.type === 'heading') {
          if (block.level === 1) {
            return (
              <h2 key={key} className='text-2xl font-semibold tracking-tight'>
                {block.text}
              </h2>
            )
          }

          if (block.level === 2) {
            return (
              <h3 key={key} className='text-lg font-semibold tracking-tight'>
                {block.text}
              </h3>
            )
          }

          return (
            <h4 key={key} className='text-base font-semibold tracking-tight'>
              {block.text}
            </h4>
          )
        }

        if (block.type === 'unordered-list') {
          return (
            <ul
              key={key}
              className='list-disc space-y-1 ps-5 text-sm text-muted-foreground'
            >
              {block.items.map((item) => (
                <li key={`${key}-${item}`}>{item}</li>
              ))}
            </ul>
          )
        }

        if (block.type === 'ordered-list') {
          return (
            <ol
              key={key}
              className='list-decimal space-y-1 ps-5 text-sm text-muted-foreground'
            >
              {block.items.map((item) => (
                <li key={`${key}-${item}`}>{item}</li>
              ))}
            </ol>
          )
        }

        return (
          <p key={key} className='text-sm leading-6 text-muted-foreground'>
            {block.text}
          </p>
        )
      })}
    </div>
  )
}

export function HelpCenter() {
  const search = route.useSearch()
  const navigate = route.useNavigate()
  const searchQuery = search.q?.trim() ?? ''
  const activeCategory = search.category?.trim() ?? ''

  const groupedArticles = useMemo(() => {
    const grouped = new Map<string, HelpCenterArticle[]>()

    for (const article of helpCenterArticles) {
      const existing = grouped.get(article.category) ?? []
      grouped.set(article.category, [...existing, article])
    }

    return Array.from(grouped.entries())
  }, [])

  const categories = useMemo(
    () =>
      Array.from(
        new Set(helpCenterArticles.map((article) => article.category))
      ),
    []
  )

  const filteredArticles = useMemo(() => {
    const query = searchQuery.toLowerCase()

    return helpCenterArticles.filter((article) => {
      const matchesCategory =
        activeCategory === '' || article.category === activeCategory
      const matchesQuery =
        query === '' ||
        [article.title, article.summary, article.category].some((value) =>
          value.toLowerCase().includes(query)
        )

      return matchesCategory && matchesQuery
    })
  }, [activeCategory, searchQuery])

  const selectedArticle =
    filteredArticles.find((article) => article.id === search.article) ??
    filteredArticles[0] ??
    helpCenterArticles[0]

  const updateSearch = (next: {
    article?: string
    category?: string
    q?: string
  }) => {
    void navigate({
      search: (prev) => ({
        ...prev,
        ...next,
      }),
      replace: next.article === undefined,
    })
  }

  return (
    <>
      <Header fixed>
        <Search />
        <HeaderTitle
          title='Help Center'
          description='Browse in-app guides, section walkthroughs, and Cabinet workflow references.'
          icon={BookOpenText}
          testId='help-center-header-title'
          iconTestId='help-center-page-icon'
        />
        <div
          className='ms-auto flex items-center space-x-4'
          data-header-title-avoid='true'
        >
          <LanguageSwitch />
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>

      <Main className='space-y-6'>
        <div className='space-y-2'>
          <h1 className='text-2xl font-bold tracking-tight'>Help Center</h1>
          <p className='text-muted-foreground'>
            Browse in-app guides, section walkthroughs, and shared Cabinet
            workflow references.
          </p>
        </div>

        <div className='grid gap-4 md:grid-cols-3'>
          <Card data-testid='help-center-library-summary'>
            <CardHeader className='pb-3'>
              <CardTitle className='flex items-center gap-2 text-base'>
                <LibraryBig className='h-4 w-4' />
                Articles available
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className='text-3xl font-semibold'>
                {helpCenterArticles.length}
              </div>
              <p className='mt-2 text-sm text-muted-foreground'>
                Help articles are grouped into Getting Started, Sections, and
                Reference guides.
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className='pb-3'>
              <CardTitle className='flex items-center gap-2 text-base'>
                <BookOpenText className='h-4 w-4' />
                Start here
              </CardTitle>
            </CardHeader>
            <CardContent className='text-sm text-muted-foreground'>
              Begin with{' '}
              <span className='font-medium text-foreground'>
                Login and Database Setup
              </span>{' '}
              to make sure you are in the right Cabinet profile before using
              Inventory, Wishlist, or Integrations.
            </CardContent>
          </Card>

          <Card>
            <CardHeader className='pb-3'>
              <CardTitle className='flex items-center gap-2 text-base'>
                <FileText className='h-4 w-4' />
                Reader behavior
              </CardTitle>
            </CardHeader>
            <CardContent className='text-sm text-muted-foreground'>
              Select any article from the library to open its in-app reading
              panel without leaving the Help Center route.
            </CardContent>
          </Card>
        </div>

        <div className='grid gap-6 xl:grid-cols-[22rem_minmax(0,1fr)]'>
          <Card data-testid='help-center-article-library'>
            <CardHeader>
              <CardTitle>Browse articles</CardTitle>
              <CardDescription>
                The Help Center now surfaces the available article set instead
                of a placeholder-only card.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className='space-y-6'>
                <div className='space-y-3'>
                  <div className='relative'>
                    <SearchIcon className='pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground' />
                    <Input
                      aria-label='Search help articles'
                      className='ps-9'
                      data-testid='help-center-article-search'
                      placeholder='Search articles'
                      value={search.q ?? ''}
                      onChange={(event) =>
                        updateSearch({
                          article: undefined,
                          q: event.target.value,
                        })
                      }
                    />
                  </div>
                  <div
                    className='flex flex-wrap gap-2'
                    data-testid='help-center-category-nav'
                  >
                    <Button
                      type='button'
                      size='sm'
                      variant={activeCategory === '' ? 'secondary' : 'outline'}
                      onClick={() =>
                        updateSearch({ article: undefined, category: '' })
                      }
                    >
                      All
                    </Button>
                    {categories.map((category) => (
                      <Button
                        key={category}
                        type='button'
                        size='sm'
                        variant={
                          activeCategory === category ? 'secondary' : 'outline'
                        }
                        data-testid={`help-center-category-${category.toLowerCase().replace(/\s+/g, '-')}`}
                        onClick={() =>
                          updateSearch({ article: undefined, category })
                        }
                      >
                        {category}
                      </Button>
                    ))}
                  </div>
                </div>

                {groupedArticles.map(([category, articles]) => (
                  <div
                    key={category}
                    className='space-y-3'
                    hidden={
                      activeCategory !== '' && activeCategory !== category
                    }
                  >
                    <div className='flex items-center justify-between'>
                      <h2 className='text-sm font-semibold tracking-wide text-muted-foreground uppercase'>
                        {category}
                      </h2>
                      <Badge variant='outline'>{articles.length}</Badge>
                    </div>
                    <div className='space-y-2'>
                      {articles
                        .filter((article) => filteredArticles.includes(article))
                        .map((article) => {
                          const selected = article.id === selectedArticle?.id
                          return (
                            <Button
                              key={article.id}
                              type='button'
                              variant={selected ? 'secondary' : 'ghost'}
                              className='h-auto w-full items-start justify-start px-3 py-3 text-left'
                              data-testid={`help-center-article-link-${article.id}`}
                              onClick={() =>
                                updateSearch({ article: article.id })
                              }
                            >
                              <div className='space-y-1'>
                                <div className='font-medium text-foreground'>
                                  {article.title}
                                </div>
                                <div className='text-xs whitespace-normal text-muted-foreground'>
                                  {article.summary}
                                </div>
                              </div>
                            </Button>
                          )
                        })}
                    </div>
                  </div>
                ))}
                {filteredArticles.length === 0 ? (
                  <div
                    className='rounded-md border border-dashed p-4 text-sm text-muted-foreground'
                    data-testid='help-center-empty-results'
                  >
                    No help articles match the current search.
                  </div>
                ) : null}
              </div>
            </CardContent>
          </Card>

          <Card data-testid='help-center-article-viewer'>
            <CardHeader>
              <div className='flex flex-wrap items-center gap-2'>
                <Badge variant='outline'>{selectedArticle.category}</Badge>
                <Badge variant='secondary'>In-app article</Badge>
              </div>
              <CardTitle data-testid='help-center-selected-article-title'>
                {selectedArticle.title}
              </CardTitle>
              <CardDescription>{selectedArticle.summary}</CardDescription>
            </CardHeader>
            <CardContent>
              <ScrollArea className='h-[32rem] pe-4'>
                <MarkdownArticle article={selectedArticle} />
              </ScrollArea>
            </CardContent>
          </Card>
        </div>
      </Main>
    </>
  )
}
