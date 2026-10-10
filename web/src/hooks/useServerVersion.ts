import { useQuery } from '@tanstack/react-query'
import { request } from '../lib/api'

export function useServerVersion() {
  const query = useQuery({
    queryKey: ['version'],
    queryFn: () => request<{ version: string }>('/version'),
    staleTime: Infinity,
    retry: false,
  })
  const version = query.data?.version ?? ''
  return {
    version,
    // Release binaries and images are tagged v<version>; local builds report "dev".
    releaseTag: version && version !== 'dev' ? `v${version}` : '',
  }
}
