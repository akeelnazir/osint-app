import { exportApi } from '@/lib/api-services'
import { useAuth } from '@/lib/auth'

export function ExportTab({ caseId }: { caseId: number }) {
  const { user } = useAuth()

  const downloadPDF = () => {
    const url = exportApi.casePDF(caseId)
    const token = localStorage.getItem('access_token')
    // Fetch with auth header, then trigger download.
    fetch(url, { headers: { Authorization: `Bearer ${token}` } })
      .then((r) => r.blob())
      .then((blob) => {
        const a = document.createElement('a')
        a.href = URL.createObjectURL(blob)
        a.download = `case-${caseId}-report.pdf`
        a.click()
        URL.revokeObjectURL(a.href)
      })
  }

  const downloadCSV = () => {
    const url = exportApi.evidenceCSV(caseId)
    const token = localStorage.getItem('access_token')
    fetch(url, { headers: { Authorization: `Bearer ${token}` } })
      .then((r) => r.blob())
      .then((blob) => {
        const a = document.createElement('a')
        a.href = URL.createObjectURL(blob)
        a.download = `case-${caseId}-evidence.csv`
        a.click()
        URL.revokeObjectURL(a.href)
      })
  }

  return (
    <div className="space-y-4">
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="font-semibold mb-3">Export Case Report</h2>
        <p className="text-sm text-slate-600 mb-4">
          Generate a PDF report containing the case summary, evidence list, map snapshot, and timeline.
        </p>
        <button onClick={downloadPDF} className="bg-brand-600 text-white px-4 py-2 rounded hover:bg-brand-700">
          Download PDF Report
        </button>
      </div>

      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="font-semibold mb-3">Export Evidence CSV</h2>
        <p className="text-sm text-slate-600 mb-4">
          Download all evidence items for this case as a CSV file.
        </p>
        <button onClick={downloadCSV} className="bg-brand-600 text-white px-4 py-2 rounded hover:bg-brand-700">
          Download CSV
        </button>
      </div>

      {user && (
        <div className="text-xs text-slate-400">
          Signed in as {user.username}. Exported reports include all evidence you have access to.
        </div>
      )}
    </div>
  )
}
