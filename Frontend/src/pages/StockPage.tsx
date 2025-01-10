import { useEffect, useState } from 'react'
import { useParams, useLocation, useNavigate } from 'react-router-dom'
import SearchBar from '../components/SearchBar'
import { DarkModeToggle } from '../components/DarkModeToggle'
import { Copy, Download } from 'lucide-react'

interface LocationState {
  cik: string
  data: FinancialStatement[]
  companyName: string
}

interface FinancialStatement {
  _id: string
  cik: string
  financialStatementType: string
  data: [string, ...string[]][]
}

export default function StockPage() {
  const { ticker } = useParams<{ ticker: string }>()
  const location = useLocation()
  const navigate = useNavigate()
  const [data, setData] = useState<FinancialStatement[]>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [selectedType, setSelectedType] = useState<'All' | 'IS' | 'BS' | 'CF'>('All')

  useEffect(() => {
    const state = location.state as LocationState
    if (!state?.data) {
      navigate('/')
      return
    }

    setData(state.data)
    setLoading(false)
  }, [location.state, navigate])

  const handleCopyToClipboard = async (rows: [string, ...string[]][], dates: string[]) => {
    try {
      const headerRow = ['Metric', ...dates].join('\t')
      const formattedRows = rows.map(row => row.join('\t'))
      const tableText = [headerRow, ...formattedRows].join('\n')
      
      await navigator.clipboard.writeText(tableText)
      alert('Table copied to clipboard!')
    } catch (error) {
      console.error('Failed to copy:', error)
      alert('Failed to copy to clipboard')
    }
  }

  const handleDownload = (rows: [string, ...string[]][], dates: string[], statementType: string) => {
    try {
      const headerRow = ['Metric', ...dates].join(',')
      const formattedRows = rows.map(row => {
        return row.map(cell => {
          return cell.includes(',') ? `"${cell}"` : cell
        }).join(',')
      })
      
      const csvContent = [headerRow, ...formattedRows].join('\n')
      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
      const link = document.createElement('a')
      const url = URL.createObjectURL(blob)
      
      link.setAttribute('href', url)
      link.setAttribute('download', `${ticker}_${statementType}_statement.csv`)
      document.body.appendChild(link)
      link.click()
      
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Failed to download:', error)
      alert('Failed to download CSV file')
    }
  }

  // Helper function to format dates
  const formatDate = (dateStr: string) => {
    if (!dateStr) return ''
    return `${dateStr.slice(0, 4)}-${dateStr.slice(4, 6)}-${dateStr.slice(6, 8)}`
  }

  // Helper function to get column headers (dates) from data
  const getColumnDates = (data: [string, ...string[]][]) => {
    const reportPeriodRow = data.find(row => row[0] === 'reportPeriod')
    if (!reportPeriodRow) return []
    return reportPeriodRow.slice(1).map(date => formatDate(date)).filter(Boolean)
  }

  // Helper function to get financial data rows
  const getFinancialRows = (data: [string, ...string[]][]) => {
    return data.filter(row => 
      row[0] && 
      !['accessionNumber', 'form', 'reportDate', 'denomination', 'reportPeriod', 'reportDurationInMonths', 'separator'].includes(row[0]) &&
      row.some((cell, index) => index > 0 && cell)
    )
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-destructive text-lg">Error: {error}</div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background relative">
      <div className="absolute top-6 right-6 z-10">
        <DarkModeToggle />
      </div>
      <div className="p-4 sm:p-8 max-w-7xl mx-auto">
        <div className="mb-8">
          <SearchBar />
        </div>

        <div className="space-y-8">
          <div>
            <h1 className="text-4xl font-bold mb-3 text-foreground">Financial Statements</h1>
            <p className="text-lg text-muted-foreground mb-6">
              {ticker} - {(location.state as LocationState)?.companyName}
            </p>
            <div className="flex flex-wrap gap-3 mb-8">
              {['All', 'Income Statement', 'Balance Sheet', 'Cash Flow Statement'].map((type) => (
                <button
                  key={type}
                  onClick={() => setSelectedType(
                    type === 'Income Statement' ? 'IS' :
                    type === 'Balance Sheet' ? 'BS' :
                    type === 'Cash Flow Statement' ? 'CF' : 'All'
                  )}
                  className={`px-4 py-2 rounded-lg transition-colors duration-200 font-medium ${
                    (selectedType === 'All' && type === 'All') ||
                    (selectedType === 'IS' && type === 'Income Statement') ||
                    (selectedType === 'BS' && type === 'Balance Sheet') ||
                    (selectedType === 'CF' && type === 'Cash Flow Statement')
                      ? 'bg-primary text-primary-foreground shadow-md'
                      : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
                  }`}
                >
                  {type}
                </button>
              ))}
            </div>
          </div>

          {data
            .sort((a, b) => {
              const order = { IS: 1, BS: 2, CF: 3 }
              return (order[a.financialStatementType as keyof typeof order] || 0) - 
                     (order[b.financialStatementType as keyof typeof order] || 0)
            })
            .filter(statement => 
              selectedType === 'All' || statement.financialStatementType === selectedType
            )
            .map((statement) => {
              const dates = getColumnDates(statement.data)
              const rows = getFinancialRows(statement.data)

              return (
                <div key={statement._id} className="space-y-4 bg-card rounded-xl p-6 shadow-sm">
                  <div className="flex flex-col sm:flex-row sm:items-center gap-4 justify-between">
                    <h2 className="text-2xl font-semibold text-foreground">
                      {statement.financialStatementType === 'BS' ? 'Balance Sheet' : 
                       statement.financialStatementType === 'IS' ? 'Income Statement' : 
                       'Cash Flow Statement'}
                    </h2>
                    <div className="flex gap-3">
                      <button
                        onClick={() => handleCopyToClipboard(rows, dates)}
                        className="px-3 py-2 text-sm rounded-lg bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors duration-200 flex items-center gap-2 font-medium"
                        title="Copy to clipboard"
                      >
                        <Copy className="w-4 h-4" />
                        Copy
                      </button>
                      <button
                        onClick={() => handleDownload(rows, dates, statement.financialStatementType)}
                        className="px-3 py-2 text-sm rounded-lg bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors duration-200 flex items-center gap-2 font-medium"
                        title="Download as CSV"
                      >
                        <Download className="w-4 h-4" />
                        Download
                      </button>
                    </div>
                  </div>
                  <div className="rounded-lg border bg-card overflow-x-auto">
                    <table className="w-full">
                      <thead>
                        <tr className="bg-muted/50">
                          <th className="py-4 px-6 text-left text-sm font-semibold text-muted-foreground border-b">Metric</th>
                          {dates.map((date, i) => (
                            <th key={i} className="py-4 px-6 text-left text-sm font-semibold text-muted-foreground border-b">
                              {date}
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        {rows.map((row, rowIndex) => (
                          <tr key={rowIndex} className="hover:bg-muted/50 transition-colors duration-200">
                            <td className="py-4 px-6 border-b text-sm font-medium whitespace-normal text-foreground">
                              {row[0]}
                            </td>
                            {row.slice(1).map((value, i) => (
                              <td key={i} className="py-4 px-6 border-b text-sm text-right text-muted-foreground">
                                {value || '-'}
                              </td>
                            ))}
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )
            })}
        </div>
      </div>
    </div>
  )
}
