import SearchBar from '../components/SearchBar'
import { DarkModeToggle } from '../components/DarkModeToggle'

export default function HomePage() {
  return (
    <div className="min-h-screen bg-background relative">
      <div className="absolute top-4 right-4">
        <DarkModeToggle />
      </div>
      <div className="flex flex-col items-center justify-center min-h-screen">
        <h1 className="text-4xl font-bold mb-8">Stock Search</h1>
        <SearchBar />
      </div>
    </div>
  )
}
