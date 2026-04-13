import Nav from '@/components/Nav'
import Hero from '@/components/Hero'
import { getEvents, getCategories, getJourneys, getMerchItems } from '@/lib/api'

export default async function Home() {
  // Fetch all data in parallel on the server
  await Promise.all([
    getEvents().catch(() => []),
    getCategories().catch(() => []),
    getJourneys().catch(() => []),
    getMerchItems().catch(() => []),
  ])

  return (
    <>
      <Nav />
      <main>
        <Hero />
      </main>
    </>
  )
}
