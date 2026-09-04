import { useQuery } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'
import L from 'leaflet'
import 'leaflet-draw'
import { evidenceApi } from '@/lib/api-services'
import { queryKeys } from '@/lib/queryKeys'

export function MapTab({ caseId }: { caseId: number }) {
  const { data, isLoading } = useQuery({
    queryKey: queryKeys.evidenceGeoJSON(caseId),
    queryFn: () => evidenceApi.geojson(caseId),
  })

  const mapRef = useRef<L.Map | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const layerRef = useRef<L.LayerGroup | null>(null)

  useEffect(() => {
    if (!containerRef.current || mapRef.current) return
    const map = L.map(containerRef.current).setView([0, 0], 2)
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; OpenStreetMap contributors',
    }).addTo(map)
    layerRef.current = L.layerGroup().addTo(map)
    mapRef.current = map

    // Draw control for filtering (leaflet-draw extends L at runtime).
    // We use a loose cast to avoid type friction with the leaflet-draw plugin.
    const DrawControl = (L as any).Control.Draw
    if (DrawControl) {
      const drawControl = new DrawControl({
        draw: { polygon: true, circle: true, rectangle: true, polyline: false, marker: false, circlemarker: false },
        edit: { featureGroup: layerRef.current },
      })
      map.addControl(drawControl)
    }

    return () => {
      map.remove()
      mapRef.current = null
    }
  }, [])

  useEffect(() => {
    if (!data || !layerRef.current || !mapRef.current) return
    layerRef.current.clearLayers()
    if (data.features.length === 0) return

    const bounds = L.latLngBounds([])
    for (const f of data.features) {
      const [lng, lat] = f.geometry.coordinates
      const marker = L.marker([lat, lng])
      marker.bindPopup(`<strong>${f.properties.title}</strong><br/>Type: ${f.properties.type}`)
      marker.addTo(layerRef.current)
      bounds.extend([lat, lng])
    }
    if (bounds.isValid()) {
      mapRef.current.fitBounds(bounds, { padding: [50, 50] })
    }
  }, [data])

  if (isLoading) return <div className="text-slate-500">Loading map...</div>

  return (
    <div className="bg-white rounded-lg shadow p-2">
      <div ref={containerRef} style={{ height: '500px' }} />
      {data && data.features.length === 0 && (
        <div className="text-center text-slate-500 py-2">No geolocated evidence.</div>
      )}
    </div>
  )
}
