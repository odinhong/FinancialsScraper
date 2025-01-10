import { useState, useEffect, KeyboardEvent, FormEvent } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import companyData from '../SEC-files/company_tickers.json'

// Type for company data
interface Company {
  cik_str: number
  ticker: string
  title: string
}

export default function SearchBar() {
  const [searchTerm, setSearchTerm] = useState('')
  const [suggestions, setSuggestions] = useState<Company[]>([])
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const isHomePage = location.pathname === '/'

  // Convert the object to an array for easier filtering
  const companies = Object.values(companyData) as Company[]

  const formatCIK = (cik: number): string => {
    return cik.toString().padStart(10, '0')
  }

  const handleSearch = (input: string) => {
    setSearchTerm(input)
    setSelectedIndex(-1)
    
    if (input.trim() === '') {
      setSuggestions([])
      return
    }

    const filteredCompanies = companies.filter(company => {
      const searchLower = input.toLowerCase()
      return (
        company.ticker.toLowerCase().includes(searchLower) ||
        company.title.toLowerCase().includes(searchLower)
      )
    }).slice(0, 5) // Limit to 5 suggestions

    setSuggestions(filteredCompanies)
  }

  const handleNavigate = async (company: Company) => {
    const formattedCIK = formatCIK(company.cik_str)
    
    try {
      setLoading(true)
      const response = await fetch(`http://localhost:3000/api/${formattedCIK}`, {
        method: 'GET',
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        }
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`Failed to fetch company data: ${response.status} ${response.statusText}${errorText ? ` - ${errorText}` : ''}`)
      }

      const data = await response.json()
      
      // Validate the data structure
      if (!Array.isArray(data)) {
        throw new Error('Invalid data format received from server')
      }

      navigate(`/stock/${company.ticker}`, { 
        state: { 
          cik: formattedCIK,
          data: data,
          companyName: company.title
        },
        replace: !isHomePage
      })
      setSearchTerm('')
      setSuggestions([])
      
    } catch (err) {
      console.error('Error fetching company data:', err)
      if (err instanceof Error) {
        alert(err.message)
      } else {
        alert('An unexpected error occurred while fetching company data')
      }
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (!searchTerm.trim() || loading) return
    
    if (selectedIndex >= 0) {
      handleNavigate(suggestions[selectedIndex])
    } else if (suggestions.length > 0) {
      handleNavigate(suggestions[0])
    }
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (suggestions.length === 0) {
      if (e.key === 'Enter') {
        e.preventDefault() // Prevent form submission on Enter with no suggestions
      }
      return
    }

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        setSelectedIndex(prev => 
          prev < suggestions.length - 1 ? prev + 1 : prev
        )
        break
      case 'ArrowUp':
        e.preventDefault()
        setSelectedIndex(prev => prev > -1 ? prev - 1 : prev)
        break
      case 'Enter':
        e.preventDefault()
        if (selectedIndex >= 0 && !loading) {
          handleNavigate(suggestions[selectedIndex])
        }
        break
      case 'Escape':
        e.preventDefault()
        setSuggestions([])
        setSelectedIndex(-1)
        setSearchTerm('')
        break
    }
  }

  // Reset selected index when suggestions change
  useEffect(() => {
    setSelectedIndex(-1)
  }, [suggestions])

  return (
    <div className="w-full max-w-2xl mx-auto relative">
      <form onSubmit={handleSubmit} className="relative">
        <div className="relative">
          <input
            type="text"
            value={searchTerm}
            onChange={(e) => handleSearch(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Search by ticker or company name"
            disabled={loading}
            className={`w-full px-4 py-3 rounded-lg border bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 disabled:opacity-50 disabled:cursor-not-allowed ${
              suggestions.length > 0 ? 'rounded-b-none border-b-0' : ''
            }`}
          />
          {loading && (
            <div className="absolute right-4 top-1/2 -translate-y-1/2">
              <div className="animate-spin rounded-full h-5 w-5 border-2 border-primary border-t-transparent"></div>
            </div>
          )}
          {searchTerm.trim() && (
            <ul className={`absolute z-10 w-full bg-background border rounded-lg ${
              suggestions.length > 0 ? 'rounded-t-none border-t-0 mt-0' : 'mt-2'
            } shadow-lg max-h-[300px] overflow-y-auto`}>
              {suggestions.length > 0 ? (
                suggestions.map((company, index) => (
                  <li
                    key={index}
                    onClick={() => !loading && handleNavigate(company)}
                    className={`px-4 py-3 flex justify-between items-center cursor-pointer hover:bg-muted/50 transition-colors duration-200 ${
                      index === selectedIndex ? 'bg-muted' : ''
                    } ${index !== suggestions.length - 1 ? 'border-b border-border/50' : ''}`}
                  >
                    <span className="font-medium text-foreground">{company.ticker}</span>
                    <span className="text-sm text-muted-foreground truncate ml-4">{company.title}</span>
                  </li>
                ))
              ) : (
                <li className="px-4 py-3 text-muted-foreground text-center">No matching companies found</li>
              )}
            </ul>
          )}
        </div>
        <button
          type="submit"
          disabled={loading || !searchTerm.trim()}
          className="mt-4 w-full px-4 py-2 bg-primary text-primary-foreground rounded-lg font-medium transition-colors duration-200 hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? 'Searching...' : 'Search'}
        </button>
      </form>
    </div>
  )
}
