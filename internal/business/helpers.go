package business

// func (b *Business) toResponse(sub *domain.Subscription) *domain.SubscriptionResponse {
// 	resp := &domain.SubscriptionResponse{
// 		ID:          sub.ID,
// 		ServiceName: sub.ServiceName,
// 		Price:       sub.Price,
// 		UserID:      sub.UserID,
// 		StartDate:   domain.FormatMonthYear(sub.StartDate),
// 		CreatedAt:   sub.CreatedAt,
// 	}

// 	if sub.EndDate != nil {
// 		formatted := domain.FormatMonthYear(*sub.EndDate)
// 		resp.EndDate = &formatted
// 	}

// 	return resp
// }

// func (b *Business) toResponseList(subs []domain.Subscription) []domain.SubscriptionResponse {
// 	result := make([]domain.SubscriptionResponse, len(subs))
// 	for i, sub := range subs {
// 		result[i] = *b.toResponse(&sub)
// 	}
// 	return result
// }
