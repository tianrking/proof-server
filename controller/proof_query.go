package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/types"
	"golang.org/x/xerrors"
)

const (
	QUERY_LIMIT = 20
)

type ProofQueryRequest struct {
	Platform string   `form:"platform"`
	Identity []string `form:"identity"`
}

type ProofQueryResponse struct {
	IDs []ProofQueryResponseSingle `json:"ids"`
}

type ProofQueryResponseSingle struct {
	Persona string                          `json:"persona"`
	Proofs  []ProofQueryResponseSingleProof `json:"proofs"`
}

type ProofQueryResponseSingleProof struct {
	Platform      types.Platform `json:"platform"`
	Identity      string         `json:"identity"`
	CreatedAt     string         `json:"created_at"`
	LastCheckedAt string         `json:"last_checked_at"`
	IsValid       bool           `json:"is_valid"`
	InvalidReason string         `json:"invalid_reason"`
}

func proofQuery(c *gin.Context) {
	req := ProofQueryRequest{}
	if err := c.BindQuery(&req); err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("Param error"))
		return
	}
	if len(req.Identity) == 0 {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("Param missing"))
		return
	}
	req.Identity = strings.Split(req.Identity[0], ",")

	c.JSON(http.StatusOK, ProofQueryResponse{
		IDs: performProofQuery(req),
	})
}

func performProofQuery(req ProofQueryRequest) []ProofQueryResponseSingle {
	result := make([]ProofQueryResponseSingle, 0, 0)

	proofs := make([]model.Proof, 0, 0)
	tx := model.DB

	switch req.Platform {
	case string(types.Platforms.NextID):
		{
			tx = tx.Where("persona IN ?", req.Identity).
				Limit(QUERY_LIMIT).
				Find(&proofs)
		}
	case "":
		{ // All platform
			tx = tx.Where("identity LIKE ?", "%"+strings.ToLower(req.Identity[0])+"%")
			for i, id := range req.Identity {
				if i == 0 {
					continue
				}
				tx = tx.Or("identity LIKE ?", "%"+strings.ToLower(id)+"%")
			}

			tx = tx.Limit(QUERY_LIMIT).Find(&proofs)
		}
	default:
		{
			tx = tx.Where("platform", req.Platform).
				Where("identity LIKE ?", "%"+strings.ToLower(req.Identity[0])+"%")

			for i, id := range req.Identity {
				if i == 0 {
					continue
				}
				tx = tx.Or("identity LIKE ?", "%"+strings.ToLower(id)+"%")
			}
			tx = tx.Limit(QUERY_LIMIT).Find(&proofs)
		}
	}
	if tx.Error != nil || tx.RowsAffected == int64(0) || len(proofs) == 0 {
		return result
	}

	// proofs.group_by(&:persona)
	persona_proof_map := make(map[string][]model.Proof, 0)
	for _, p := range proofs {
		persona_proof, ok := persona_proof_map[p.Persona]
		if ok {
			persona_proof_map[p.Persona] = append(persona_proof, p)
		} else {
			persona_proof_map[p.Persona] = append(make([]model.Proof, 0, 0), p)
		}
	}

	for persona, proofs := range persona_proof_map {
		single := ProofQueryResponseSingle{
			Persona: persona,
			Proofs:  make([]ProofQueryResponseSingleProof, 0),
		}
		for _, p := range proofs {
			single.Proofs = append(single.Proofs, ProofQueryResponseSingleProof{
				Platform:      p.Platform,
				Identity:      p.Identity,
				CreatedAt:     strconv.FormatInt(p.CreatedAt.Unix(), 10),
				LastCheckedAt: strconv.FormatInt(p.LastCheckedAt.Unix(), 10),
				IsValid:       p.IsValid,
				InvalidReason: p.InvalidReason,
			})
		}
		result = append(result, single)
	}

	return result
}
