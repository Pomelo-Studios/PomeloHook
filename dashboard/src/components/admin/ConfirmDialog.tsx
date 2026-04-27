interface Props {
  message: string
  detail?: string
  onConfirm: () => void
  onCancel: () => void
}

export function ConfirmDialog({ message, detail, onConfirm, onCancel }: Props) {
  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
      <div className="bg-zinc-900 border border-zinc-700 rounded-lg p-5 max-w-md w-full mx-4 shadow-xl">
        <p className="text-zinc-200 text-sm font-medium mb-1">{message}</p>
        {detail && <p className="text-zinc-500 text-xs font-mono break-all mb-4">{detail}</p>}
        <div className="flex justify-end gap-2 mt-4">
          <button
            onClick={onCancel}
            className="text-xs px-3 py-1.5 border border-zinc-700 text-zinc-400 rounded hover:text-zinc-200 hover:border-zinc-500"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className="text-xs px-3 py-1.5 bg-red-900 text-red-200 rounded hover:bg-red-800 font-medium"
          >
            Confirm
          </button>
        </div>
      </div>
    </div>
  )
}
