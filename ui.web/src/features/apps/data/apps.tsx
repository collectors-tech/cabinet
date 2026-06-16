import { Store } from 'lucide-react'
import {
  IconTelegram,
  IconGmail,
  IconGithub,
  IconWhatsapp,
} from '@/assets/brand-icons'

const StoreLogo = () => <Store className='size-4' />

export const apps = [
  {
    name: 'eBay',
    logo: <StoreLogo />,
    connected: false,
    desc: 'Marketplace connector for scanner discovery, pricing, and matching.',
  },
  {
    name: 'Telegram',
    logo: <IconTelegram />,
    connected: false,
    desc: 'Connect with Telegram for real-time communication.',
  },
  {
    name: 'Gmail',
    logo: <IconGmail />,
    connected: true,
    desc: 'Access and manage Gmail messages effortlessly.',
  },
  {
    name: 'GitHub',
    logo: <IconGithub />,
    connected: false,
    desc: 'Streamline code management with GitHub integration.',
  },
  {
    name: 'WhatsApp',
    logo: <IconWhatsapp />,
    connected: false,
    desc: 'Easily integrate WhatsApp for direct messaging.',
  },
  {
    name: 'Mr Toys',
    logo: <StoreLogo />,
    connected: false,
    desc: 'Retail provider connector for availability and price checks.',
  },
  {
    name: 'Voglers',
    logo: <StoreLogo />,
    connected: false,
    desc: 'Retail provider connector for availability and price checks.',
  },
  {
    name: 'WA Slot Cars',
    logo: <StoreLogo />,
    connected: false,
    desc: 'Retail provider connector for availability and price checks.',
  },
  {
    name: 'Bonza Slot Cars',
    logo: <StoreLogo />,
    connected: false,
    desc: 'Retail provider connector for availability and price checks.',
  },
  {
    name: 'K & K Creative Toys',
    logo: <StoreLogo />,
    connected: false,
    desc: 'Retail provider connector for availability and price checks.',
  },
  {
    name: 'Hobbyco',
    logo: <StoreLogo />,
    connected: false,
    desc: 'Retail provider connector for availability and price checks.',
  },
  {
    name: 'Frontline Hobbies',
    logo: <StoreLogo />,
    connected: false,
    desc: 'Retail provider connector for availability and price checks.',
  },
  {
    name: 'Metro Hobbies',
    logo: <StoreLogo />,
    connected: false,
    desc: 'Retail provider connector for availability and price checks.',
  },
]
