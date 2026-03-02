type PageTitleProps = {
  title: string
}

const PageTitle = ({ title }: PageTitleProps) => {
  return <h1 className="text-gray-200 text-4xl bg-transparent font-semibold">{title}</h1>
}

export default PageTitle
