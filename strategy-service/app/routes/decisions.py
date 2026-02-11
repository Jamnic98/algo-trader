from fastapi import status, HTTPException
from fastapi.routing import APIRouter

from app.models.decision import DecisionRequest, DecisionResponse  # , BatchDecisionRequest, BatchDecisionResponse

router = APIRouter()


@router.post("/")
def make_decision(data: DecisionRequest) -> DecisionResponse:
    if data.symbol not in ["BTC-USD", "ETH-USD"]:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Symbol {data.symbol} not supported"
        )
    
    return DecisionResponse(action="hold")


# @router.post("/batch/")
# def make_batch_decisions(data: BatchDecisionRequest) -> BatchDecisionResponse:
#     return BatchDecisionResponse
