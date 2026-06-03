import type { RecommendationInput, RecommendationRoadmap } from './types'

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/+$/, '')

export async function createRecommendation(
  input: RecommendationInput,
): Promise<RecommendationRoadmap> {
  const response = await fetch(`${API_BASE_URL}/api/recommendations`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })

  if (!response.ok) {
    let message = 'Could not create recommendation.'
    try {
      const body = (await response.json()) as { error?: string | { message?: string } }
      if (typeof body.error === 'string') {
        message = body.error
      }
      if (body.error && typeof body.error === 'object' && body.error.message) {
        message = body.error.message
      }
    } catch {
      // Keep the fallback message when the backend response is not JSON.
    }
    throw new Error(message)
  }

  return (await response.json()) as RecommendationRoadmap
}
