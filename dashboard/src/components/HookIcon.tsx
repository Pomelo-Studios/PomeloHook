interface Props {
  size?: number
}

export function HookIcon({ size = 28 }: Props) {
  return (
    <div
      style={{ width: size, height: size, borderRadius: Math.round(size * 0.286) }}
      className="bg-coral flex items-center justify-center flex-shrink-0"
    >
      <svg width={size * 0.57} height={size * 0.57} viewBox="0 0 52 52" fill="none">
        <path
          d="M18 14 L18 30 Q18 40 28 40 Q38 40 38 30"
          stroke="white" strokeWidth="5" strokeLinecap="round" fill="none"
        />
        <path
          d="M33 25 L38 30 L43 25"
          stroke="white" strokeWidth="4.5" strokeLinecap="round" strokeLinejoin="round" fill="none"
        />
        <circle cx="18" cy="11" r="4" fill="white" opacity="0.9" />
      </svg>
    </div>
  )
}
