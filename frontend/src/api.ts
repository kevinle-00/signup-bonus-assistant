import type { RecommendationInput, RecommendationRoadmap } from './types'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

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
      const body = (await response.json()) as { error?: string }
      if (body.error) {
        message = body.error
      }
    } catch {
      // Keep the fallback message when the backend response is not JSON.
    }
    throw new Error(message)
  }

  return (await response.json()) as RecommendationRoadmap
}
