interface Props {
  message: string
  detail?: string
  onConfirm: () => void
  onCancel: () => void
}

export function ConfirmDialog({ message, detail, onConfirm, onCancel }: Props) {
  return (
    <div className="fixed inset-0 flex items-center justify-center z-50" style={{ background: 'rgba(0,0,0,0.6)' }}>
      <div
        className="rounded-xl p-5 max-w-md w-full mx-4 shadow-xl"
        style={{ background: 'var(--surface)', border: '1px solid var(--border)' }}
      >
        <p className="text-sm font-semibold mb-1" style={{ color: 'var(--text)' }}>{message}</p>
        {detail && (
          <p className="font-mono text-xs break-all mb-4" style={{ color: 'var(--text-2)' }}>{detail}</p>
        )}
        <div className="flex justify-end gap-2 mt-4">
          <button
            onClick={onCancel}
            className="text-xs px-3 py-[6px] rounded-lg transition-colors"
            style={{ border: '1px solid var(--border)', color: 'var(--text-2)' }}
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className="text-xs px-3 py-[6px] rounded-lg font-semibold"
            style={{ background: 'var(--err-bg)', color: 'var(--err-text)', border: '1px solid var(--selected-border)' }}
          >
            Confirm
          </button>
        </div>
      </div>
    </div>
  )
}
