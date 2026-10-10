import { KeyRound } from 'lucide-react'
import { CopyField } from '../../components/ui/CopyField'
import { useServerVersion } from '../../hooks/useServerVersion'

export function TokenReveal({ token }: { token: string }) {
  const { releaseTag } = useServerVersion()
  const tag = releaseTag || '<release-tag>'
  const command = [
    `curl -fsSLO https://github.com/z9wen/VPSentinel/releases/download/${tag}/install-agent.sh`,
    `sudo VPSENTINEL_VERSION=${tag} \\`,
    `  VPSENTINEL_SERVER_URL=${window.location.origin}/api/v1/agent/metrics \\`,
    `  VPSENTINEL_TOKEN=${token} \\`,
    '  sh install-agent.sh',
  ].join('\n')

  return (
    <div className="space-y-4">
      <div className="flex gap-3 rounded-field border border-warning/30 bg-warning/10 px-3 py-2.5 text-sm">
        <KeyRound size={16} className="mt-0.5 shrink-0" />
        <p>
          Copy this token now. The server keeps only its SHA-256 hash, so it cannot be shown again.
        </p>
      </div>
      <CopyField label="Agent token" value={token} />
      <CopyField label="Install on the host" value={command} multiline />
    </div>
  )
}
