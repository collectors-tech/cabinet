import { Check, Languages } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

type LanguageOption = {
  code: string
  label: string
}

const languages: LanguageOption[] = [
  { code: 'en', label: 'EN' },
  { code: 'ar', label: 'AR' },
]

export function LanguageSwitch() {
  const { i18n, t } = useTranslation('common')
  const current =
    languages.find((lang) => i18n.language.startsWith(lang.code)) ?? languages[0]

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button
          variant='ghost'
          size='sm'
          className='rounded-full px-3'
          data-testid='header-language-switch-trigger'
        >
          <Languages className='me-2 size-4' />
          <span>{current.label}</span>
          <span className='sr-only'>{t('language.label')}</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end'>
        {languages.map((lang) => (
          <DropdownMenuItem
            key={lang.code}
            data-testid={`header-language-option-${lang.code}`}
            onClick={() => void i18n.changeLanguage(lang.code)}
          >
            {lang.label}
            <Check
              size={14}
              className={cn('ms-auto', i18n.language !== lang.code && 'hidden')}
            />
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
