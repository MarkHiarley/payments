package apperr

// Kind classifies a domain error without committing to a transport. The HTTP
// layer maps each Kind to exactly one status code.
type Kind string

const (
	KindValidation   Kind = "validation"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindNotFound     Kind = "not_found"
	KindConflict     Kind = "conflict"
	KindPrecondition Kind = "precondition"
	KindUnavailable  Kind = "unavailable"
	KindInternal     Kind = "internal"
)

func title(kind Kind) string {
	switch kind {
	case KindValidation:
		return "Requisição inválida"
	case KindUnauthorized:
		return "Não autenticado"
	case KindForbidden:
		return "Acesso negado"
	case KindNotFound:
		return "Recurso não encontrado"
	case KindConflict:
		return "Conflito de estado"
	case KindPrecondition:
		return "Pré-condição não satisfeita"
	case KindUnavailable:
		return "Serviço indisponível"
	default:
		return "Erro interno"
	}
}
