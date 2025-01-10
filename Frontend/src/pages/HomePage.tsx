import SearchBar from '../components/SearchBar'
import { DarkModeToggle } from '../components/DarkModeToggle'

export default function HomePage() {
  return (
    <div className="min-h-screen bg-background relative">
      <div className="absolute top-6 right-6 z-10">
        <DarkModeToggle />
      </div>
      <div className="flex flex-col items-center justify-center min-h-screen p-4">
        <div className="w-full max-w-4xl mx-auto text-center space-y-8">
          <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold text-foreground tracking-tight">
            Financial Statement Search
          </h1>
          <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
            Search for companies and view their financial statements. Enter a company name or ticker symbol to get started.
          </p>
          <div className="mt-8">
            <SearchBar />
          </div>
        </div>
      </div>
    </div>
  )
}
