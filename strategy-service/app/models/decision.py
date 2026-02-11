from pydantic import BaseModel


class DecisionRequest(BaseModel):
    symbol: str


class DecisionResponse(BaseModel):
    action: str


# class BatchDecisionRequest(BaseModel):
#     requests: list[DecisionRequest]


# class BatchDecisionResponse(BaseModel):
#     action: str
