'use client'

import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

export default function SearchBar() {
  const [ticker, setTicker] = useState('')
  const navigate = useNavigate()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (ticker) {
      navigate(`/stock/${ticker}`)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex w-full max-w-sm space-x-3">
      <input
        type="text"
        value={ticker}
        onChange={(e) => setTicker(e.target.value)}
        placeholder="Enter stock ticker"
        className="flex-1 appearance-none border rounded-lg py-2 px-4 bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
      />
      <button
        type="submit"
        className="flex-shrink-0 bg-primary text-primary-foreground font-semibold py-2 px-4 rounded-lg shadow-md hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2"
      >
        Search
      </button>
    </form>
  )
}
