import { Heading, StrategyCreationWizard } from 'components'

const StrategyCreator = () => {
  return (
    <div className="space-y-8">
      <Heading title="Strategy Creator" />
      <div className="space-y-4">
        <StrategyCreationWizard />
      </div>
    </div>
  )
}

export default StrategyCreator
