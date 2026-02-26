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

const languages: LanguageOption[] = [{ code: 'en', label: 'EN' }]

export function LanguageSwitch() {
  const { i18n, t } = useTranslation('common')
  const current = languages.find((lang) => lang.code === i18n.language) ?? languages[0]

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant='ghost' size='sm' className='rounded-full px-3'>
          <Languages className='me-2 size-4' />
          <span>{current.label}</span>
          <span className='sr-only'>{t('language.label')}</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end'>
        {languages.map((lang) => (
          <DropdownMenuItem key={lang.code} onClick={() => void i18n.changeLanguage(lang.code)}>
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

