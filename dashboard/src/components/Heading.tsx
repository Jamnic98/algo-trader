import type { JSX } from 'react'

type HeadingSize = 1 | 2 | 3 | 4 | 5 | 6

type HeadingProps = {
  title: string
  as?: HeadingSize
  size?: HeadingSize
}

const sizeMap: Record<HeadingSize, string> = {
  1: 'text-4xl',
  2: 'text-3xl',
  3: 'text-2xl',
  4: 'text-xl',
  5: 'text-lg',
  6: 'text-base',
}

const Heading = ({ title, as = 1, size }: HeadingProps) => {
  const Tag = `h${as}` as keyof JSX.IntrinsicElements
  return (
    <Tag
      className={`text-gray-400 ${sizeMap[size ?? as]} bg-transparent font-semibold select-none`}
    >
      {title}
    </Tag>
  )
}

export default Heading
