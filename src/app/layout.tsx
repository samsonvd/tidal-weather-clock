import CONSTANTS from '@/lib/constants'
import './globals.css'
import Navbar from '@/components/nav-bar/Navbar'
import Providers from '@/components/Providers'
import { Analytics } from '@vercel/analytics/next'
import React from 'react'

export default function Layout({ children }: React.PropsWithChildren) {
  return (
    <html lang="en">
      <head>
        <title>{CONSTANTS.TITLE}</title>
        <link href="https://rsms.me/inter/inter.css" rel="stylesheet" />
        <meta content="width=device-width, initial-scale=1.0" name="viewport" />
      </head>
      <body>
        <div id="root">
          <Providers>
            <Navbar />
            <div className="mx-auto flex min-w-full flex-col justify-center gap-2 p-4 text-center md:min-w-0 md:p-10">
              {children}
            </div>
          </Providers>
        </div>
        <Analytics />
      </body>
    </html>
  )
}
