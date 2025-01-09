# SEC Financial Data Scraper

## 🚀 Overview
A high-performance, Go-based financial data extraction tool that automatically retrieves and processes SEC filings (10-K and 10-Q) into structured, analysis-ready formats. This project demonstrates expertise in building scalable data pipelines, working with financial data, and implementing efficient data processing algorithms.

## ✨ Key Features
- **Automated SEC Filing Retrieval**: Seamlessly fetches 10-K (annual) and 10-Q (quarterly) reports from the SEC's EDGAR database
- **Intelligent Data Extraction**: Automatically identifies and extracts key financial statements:
  - Income Statements
  - Balance Sheets
  - Cash Flow Statements
- **Data Processing Pipeline**: Transforms raw filing data into clean, structured CSV format
- **MongoDB Integration**: Implements robust data persistence with MongoDB Atlas
- **Environment Security**: Utilizes secure environment variable management for sensitive credentials

## 🛠️ Technical Stack
- **Language**: Go
- **Database**: MongoDB Atlas
- **External APIs**: SEC EDGAR
- **Dependencies Management**: Go Modules
- **Environment Management**: godotenv
- **Data Formats**: CSV, XBRL (SEC filings)

## 🏗️ Architecture
The project is organized into specialized modules:
- `fetchDataFolder/`: Handles SEC EDGAR API interactions and data retrieval
- `parseRfiles/`: Processes raw filing data
- `categorizeRfiles/`: Classifies and organizes financial statements
- `combineCSVfiles/`: Aggregates and structures data into final format
- `utilityFunctions/`: Common utilities and helper functions


## 🔧 Technical Implementation Highlights
- **Concurrent Processing**: Implements Go's concurrency patterns for efficient data processing
- **Error Handling**: Robust error management and logging system
- **Data Validation**: Comprehensive validation of financial data integrity
- **Modular Design**: Clean architecture with separate concerns for maintainability
- **Database Integration**: Efficient MongoDB operations with proper connection management
