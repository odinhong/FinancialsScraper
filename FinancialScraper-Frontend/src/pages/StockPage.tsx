import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import SearchBar from '../components/SearchBar'
import { DarkModeToggle } from '../components/DarkModeToggle'
import { Copy, Download } from 'lucide-react'

interface FinancialStatement {
  _id: string
  cik: string
  financialStatementType: string
  data: [string, ...string[]][]
}

async function getStockData(ticker: string) {
  console.log('Fetching data for ticker:', ticker)
  const res = await fetch(`http://localhost:3000/api/${ticker}`, { cache: 'no-store' })
  if (!res.ok) {
    console.error('API Error:', res.status, res.statusText)
    throw new Error('Failed to fetch data')
  }
  const data = await res.json()
  console.log('Raw API Response:', JSON.stringify(data, null, 2))
  return data
}

export default function StockPage() {
  const { ticker } = useParams<{ ticker: string }>()
  const [data, setData] = useState<FinancialStatement[]>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [selectedType, setSelectedType] = useState<'All' | 'IS' | 'BS' | 'CF'>('All')

  const handleCopyToClipboard = async (rows: [string, ...string[]][], dates: string[]) => {
    try {
      // Create header row with dates
      const headerRow = ['Metric', ...dates].join('\t');
      
      // Format each data row
      const formattedRows = rows.map(row => {
        return row.join('\t');
      });
      
      // Combine all rows with newlines
      const tableText = [headerRow, ...formattedRows].join('\n');
      
      // Copy to clipboard
      await navigator.clipboard.writeText(tableText);
      
      // Optional: Show a success message
      alert('Table copied to clipboard!');
    } catch (error) {
      console.error('Failed to copy:', error);
      alert('Failed to copy to clipboard');
    }
  }

  const handleDownload = (rows: [string, ...string[]][], dates: string[], statementType: string) => {
    try {
      // Create header row with dates
      const headerRow = ['Metric', ...dates].join(',');
      
      // Format each data row, properly escaping cells that contain commas
      const formattedRows = rows.map(row => {
        return row.map(cell => {
          // If cell contains comma, quote it
          return cell.includes(',') ? `"${cell}"` : cell;
        }).join(',');
      });
      
      // Combine all rows with newlines
      const csvContent = [headerRow, ...formattedRows].join('\n');
      
      // Create blob and download link
      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
      const link = document.createElement('a');
      const url = URL.createObjectURL(blob);
      
      link.setAttribute('href', url);
      link.setAttribute('download', `${ticker}_${statementType}_statement.csv`);
      document.body.appendChild(link);
      link.click();
      
      // Cleanup
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to download:', error);
      alert('Failed to download CSV file');
    }
  }

  useEffect(() => {
    let isMounted = true;

    async function fetchData() {
      if (!ticker) {
        setLoading(false);
        return;
      }

      try {
        console.log('Starting data fetch for ticker:', ticker)
        const stockData = await getStockData(ticker)
        console.log('Processed stock data:', stockData)
        if (isMounted) {
          console.log('Setting data in state:', stockData)
          setData(stockData)
        }
      } catch (err) {
        if (isMounted) {
          setError(err instanceof Error ? err.message : 'Failed to fetch data')
        }
      } finally {
        if (isMounted) {
          setLoading(false)
        }
      }
    }

    fetchData()
    
    return () => {
      isMounted = false;
    }
  }, [ticker])

  // Helper function to format dates
  const formatDate = (dateStr: string) => {
    if (!dateStr) return '';
    return `${dateStr.slice(0, 4)}-${dateStr.slice(4, 6)}-${dateStr.slice(6, 8)}`;
  };

  // Helper function to get column headers (dates) from data
  const getColumnDates = (data: [string, ...string[]][]) => {
    const reportPeriodRow = data.find(row => row[0] === 'reportPeriod');
    if (!reportPeriodRow) return [];
    return reportPeriodRow.slice(1).map(date => formatDate(date)).filter(Boolean);
  };

  // Helper function to get financial data rows
  const getFinancialRows = (data: [string, ...string[]][]) => {
    return data.filter(row => 
      row[0] && 
      !['accessionNumber', 'form', 'reportDate', 'denomination', 'reportPeriod', 'reportDurationInMonths', 'separator'].includes(row[0]) &&
      row.some((cell, index) => index > 0 && cell)
    );
  };

  return (
    <div className="min-h-screen bg-background relative">
      <div className="absolute top-4 right-4">
        <DarkModeToggle />
      </div>
      <div className="p-8 max-w-6xl mx-auto">
        {loading ? (
          <div className="flex items-center justify-center min-h-[calc(100vh-4rem)]">
            <div className="text-lg">Loading...</div>
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center min-h-[calc(100vh-4rem)]">
            <div className="text-lg text-destructive mb-4">Error: {error}</div>
            <SearchBar />
          </div>
        ) : data.length === 0 ? (
          <div className="flex flex-col items-center justify-center min-h-[calc(100vh-4rem)]">
            <div className="text-lg mb-4">No data found for ticker: {ticker}</div>
            <SearchBar />
          </div>
        ) : (
          <>
            <div className="mb-8">
              <SearchBar />
            </div>
            <div className="space-y-8">
              <div>
                <h1 className="text-3xl font-bold mb-2">Financial Statements</h1>
                <p className="text-muted-foreground mb-4">CIK: {data[0].cik}</p>
                <div className="flex space-x-4 mb-6">
                  {['All', 'Income Statement', 'Balance Sheet', 'Cash Flow Statement'].map((type) => (
                    <button
                      key={type}
                      onClick={() => setSelectedType(
                        type === 'Income Statement' ? 'IS' :
                        type === 'Balance Sheet' ? 'BS' :
                        type === 'Cash Flow Statement' ? 'CF' : 'All'
                      )}
                      className={`px-4 py-2 rounded-md ${
                        (selectedType === 'All' && type === 'All') ||
                        (selectedType === 'IS' && type === 'Income Statement') ||
                        (selectedType === 'BS' && type === 'Balance Sheet') ||
                        (selectedType === 'CF' && type === 'Cash Flow Statement')
                          ? 'bg-primary text-primary-foreground'
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
                  const order = { IS: 1, BS: 2, CF: 3 };
                  return (order[a.financialStatementType as keyof typeof order] || 0) - 
                         (order[b.financialStatementType as keyof typeof order] || 0);
                })
                .filter(statement => 
                  selectedType === 'All' || statement.financialStatementType === selectedType
                )
                .map((statement) => {
                const dates = getColumnDates(statement.data);
                const rows = getFinancialRows(statement.data);

                return (
                  <div key={statement._id} className="space-y-4">
                    <div className="flex items-center gap-4">
                      <h2 className="text-2xl font-semibold">
                        {statement.financialStatementType === 'BS' ? 'Balance Sheet' : 
                         statement.financialStatementType === 'IS' ? 'Income Statement' : 
                         'Cash Flow Statement'}
                      </h2>
                      <div className="flex gap-2">
                        <button
                          onClick={() => handleCopyToClipboard(rows, dates)}
                          className="px-2 py-1 text-sm rounded-md bg-secondary text-secondary-foreground hover:bg-secondary/80 flex items-center gap-1"
                          title="Copy to clipboard"
                        >
                          <Copy className="w-4 h-4" />
                          Copy
                        </button>
                        <button
                          onClick={() => handleDownload(rows, dates, statement.financialStatementType)}
                          className="px-2 py-1 text-sm rounded-md bg-secondary text-secondary-foreground hover:bg-secondary/80 flex items-center gap-1"
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
                          <tr>
                            <th className="py-3 px-4 text-left text-sm font-medium text-muted-foreground border-b">Metric</th>
                            {dates.map((date, i) => (
                              <th key={i} className="py-3 px-4 text-left text-sm font-medium text-muted-foreground border-b">
                                {date}
                              </th>
                            ))}
                          </tr>
                        </thead>
                        <tbody>
                          {rows.map((row, rowIndex) => (
                            <tr key={rowIndex} className="hover:bg-muted/50">
                              <td className="py-3 px-4 border-b text-sm font-medium whitespace-normal">
                                {row[0]}
                              </td>
                              {row.slice(1).map((value, i) => (
                                <td key={i} className="py-3 px-4 border-b text-sm text-right">
                                  {value || '-'}
                                </td>
                              ))}
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                );
              })}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
