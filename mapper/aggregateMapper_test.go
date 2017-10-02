package mapper

import (
	"testing"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/assert"
	"github.com/pkg/errors"
	"strings"
)

func TestAggregateMapperMap_Ok(t *testing.T) {
	mockResolver := new(model.MockIResolver)
	mockValidator := new(model.MockCPHValidator)
	mockContentMapper := new(model.MockCPHMapper)
	mockCompContentMapper := new(model.MockCPHMapper)

	givenMethodeCPH := &model.MethodeContentPlaceholder{}

	expectedUppContents := []model.UppContent{
		&model.UppContentPlaceholder{
			UppCoreContent: model.UppCoreContent{
				UUID:             "512c1f3d-e48c-4618-863c-94bc9d913b9b",
				PublishReference: "tid_test123",
				LastModified:     "2017-05-15T15:54:32.166Z",
				ContentURI:       "",
				IsMarkedDeleted:  false,
			},
		},
		&model.UppComplementaryContent{
			UppCoreContent: model.UppCoreContent{
				UUID:             "512c1f3d-e48c-4618-863c-94bc9d913b9b",
				PublishReference: "tid_test123",
				LastModified:     "2017-05-15T15:54:32.166Z",
				ContentURI:       "",
				IsMarkedDeleted:  false,
			},
		},
	}

	mockValidator.On("Validate",
		mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true })).
		Return(nil)

	mockResolver.On("ResolveIdentifier",
		mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }),
		mock.MatchedBy(func(uuid string) bool { return true }),
		mock.MatchedBy(func(tid string) bool { return true }),
		mock.MatchedBy(func(lmd string) bool { return true })).
		Return("512c1f3d-e48c-4618-863c-94bc9d913b9b", nil)

	mockContentMapper.On("MapContentPlaceholder",
		mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }),
		mock.MatchedBy(func(uuid string) bool { return true }),
		mock.MatchedBy(func(tid string) bool { return true }),
		mock.MatchedBy(func(lmd string) bool { return true })).
		Return([]model.UppContent{expectedUppContents[0]}, nil)

	mockCompContentMapper.On("MapContentPlaceholder",
		mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }),
		mock.MatchedBy(func(uuid string) bool { return true }),
		mock.MatchedBy(func(tid string) bool { return true }),
		mock.MatchedBy(func(lmd string) bool { return true })).
		Return([]model.UppContent{expectedUppContents[1]}, nil)

	aggregateMapper := NewAggregateCPHMapper(mockResolver, mockValidator, []CPHMapper{mockContentMapper, mockCompContentMapper})

	actualUppContents, err := aggregateMapper.MapContentPlaceholder(givenMethodeCPH, "tid_test123", "2017-05-15T15:54:32.166Z")
	assert.NoError(t, err, "No error should be thrown for correct mapping.")

	assert.Equal(t, expectedUppContents[0], actualUppContents[0])
	assert.Equal(t, expectedUppContents[1], actualUppContents[1])
}

func TestAggregateMapperValidationError_ThrowsException(t *testing.T) {
	mockResolver := new(model.MockIResolver)
	mockValidator := new(model.MockCPHValidator)
	mockContentMapper := new(model.MockCPHMapper)
	mockCompContentMapper := new(model.MockCPHMapper)

	givenMethodeCPH := &model.MethodeContentPlaceholder{}

	mockValidator.On("Validate",
		mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true })).
		Return(errors.New("Some validation error"))

	aggregateMapper := NewAggregateCPHMapper(mockResolver, mockValidator, []CPHMapper{mockContentMapper, mockCompContentMapper})

	_, err := aggregateMapper.MapContentPlaceholder(givenMethodeCPH, "tid_test123", "2017-05-15T15:54:32.166Z")
	assert.Error(t, err, "An error should be thrown for validation error.")
}

func TestAggregateMapperIResolverUUID_IsSet(t *testing.T) {
	mockResolver := new(model.MockIResolver)
	mockValidator := new(model.MockCPHValidator)
	mockContentMapper := new(model.MockCPHMapper)
	mockCompContentMapper := new(model.MockCPHMapper)

	givenMethodeCPH := &model.MethodeContentPlaceholder{
		UUID: "cdac1f3d-e48c-4618-863c-94bc9d913b9b",
		Attributes: model.Attributes{
			Category:  "blog",
			ServiceId: "1111",
			RefField:  "7777",
		},
	}

	expectedUppContents := []model.UppContent{
		&model.UppContentPlaceholder{
			UppCoreContent: model.UppCoreContent{
				UUID:             "abac1f3d-e48c-4618-863c-94bc9d913b9b",
				PublishReference: "tid_test123",
				LastModified:     "2017-05-15T15:54:32.166Z",
				ContentURI:       "",
				IsMarkedDeleted:  false,
			},
		},
		&model.UppComplementaryContent{
			UppCoreContent: model.UppCoreContent{
				UUID:             "abac1f3d-e48c-4618-863c-94bc9d913b9b",
				PublishReference: "tid_test123",
				LastModified:     "2017-05-15T15:54:32.166Z",
				ContentURI:       "",
				IsMarkedDeleted:  false,
			},
		},
	}

	mockValidator.On("Validate",
		mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true })).
		Return(nil)

	mockResolver.On("ResolveIdentifier",
		mock.MatchedBy(func(uuid string) bool { return true }),
		mock.MatchedBy(func(tid string) bool { return true }),
		mock.MatchedBy(func(lmd string) bool { return true })).
		Return("abac1f3d-e48c-4618-863c-94bc9d913b9b", nil)

	mockContentMapper.On("MapContentPlaceholder",
		mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }),
		mock.MatchedBy(func(uuid string) bool { return strings.EqualFold(uuid, "abac1f3d-e48c-4618-863c-94bc9d913b9b") }),
		mock.MatchedBy(func(tid string) bool { return true }),
		mock.MatchedBy(func(lmd string) bool { return true })).
		Return([]model.UppContent{expectedUppContents[0]}, nil)

	mockCompContentMapper.On("MapContentPlaceholder",
		mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }),
		mock.MatchedBy(func(uuid string) bool { return strings.EqualFold(uuid, "abac1f3d-e48c-4618-863c-94bc9d913b9b") }),
		mock.MatchedBy(func(tid string) bool { return true }),
		mock.MatchedBy(func(lmd string) bool { return true })).
		Return([]model.UppContent{expectedUppContents[1]}, nil)

	aggregateMapper := NewAggregateCPHMapper(mockResolver, mockValidator, []CPHMapper{mockContentMapper, mockCompContentMapper})

	actualUppContents, err := aggregateMapper.MapContentPlaceholder(givenMethodeCPH, "tid_test123", "2017-05-15T15:54:32.166Z")
	assert.NoError(t, err, "No error should be thrown for correct mapping.")

	assert.Equal(t, "abac1f3d-e48c-4618-863c-94bc9d913b9b", actualUppContents[0].GetUUID())
	assert.Equal(t, "abac1f3d-e48c-4618-863c-94bc9d913b9b", actualUppContents[1].GetUUID())
}

func TestAggregateMapperNotBlog_NoUUIDResolved(t *testing.T) {
	mockResolver := new(model.MockIResolver)
	mockValidator := new(model.MockCPHValidator)
	mockContentMapper := new(model.MockCPHMapper)
	mockCompContentMapper := new(model.MockCPHMapper)

	givenMethodeCPH := &model.MethodeContentPlaceholder{
		UUID: "cdac1f3d-e48c-4618-863c-94bc9d913b9b",
		Attributes: model.Attributes{
			Category:  "not-a-blog-category",
			ServiceId: "1111",
			RefField:  "7777",
		},
	}

	expectedUppContents := []model.UppContent{
		&model.UppContentPlaceholder{
			UppCoreContent: model.UppCoreContent{
				UUID:             "cdac1f3d-e48c-4618-863c-94bc9d913b9b",
				PublishReference: "tid_test123",
				LastModified:     "2017-05-15T15:54:32.166Z",
				ContentURI:       "",
				IsMarkedDeleted:  false,
			},
		},
		&model.UppComplementaryContent{
			UppCoreContent: model.UppCoreContent{
				UUID:             "cdac1f3d-e48c-4618-863c-94bc9d913b9b",
				PublishReference: "tid_test123",
				LastModified:     "2017-05-15T15:54:32.166Z",
				ContentURI:       "",
				IsMarkedDeleted:  false,
			},
		},
	}

	mockValidator.On("Validate",
		mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true })).
		Return(nil)

	mockResolver.On("ResolveIdentifier",
		mock.MatchedBy(func(uuid string) bool { return true }),
		mock.MatchedBy(func(tid string) bool { return true }),
		mock.MatchedBy(func(lmd string) bool { return true })).
		Return("abac1f3d-e48c-4618-863c-94bc9d913b9b", nil)

	mockContentMapper.On("MapContentPlaceholder",
		mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }),
		mock.MatchedBy(func(uuid string) bool { return strings.EqualFold(uuid, "") }),
		mock.MatchedBy(func(tid string) bool { return true }),
		mock.MatchedBy(func(lmd string) bool { return true })).
		Return([]model.UppContent{expectedUppContents[0]}, nil)

	mockCompContentMapper.On("MapContentPlaceholder",
		mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }),
		mock.MatchedBy(func(uuid string) bool { return strings.EqualFold(uuid, "") }),
		mock.MatchedBy(func(tid string) bool { return true }),
		mock.MatchedBy(func(lmd string) bool { return true })).
		Return([]model.UppContent{expectedUppContents[1]}, nil)

	aggregateMapper := NewAggregateCPHMapper(mockResolver, mockValidator, []CPHMapper{mockContentMapper, mockCompContentMapper})

	actualUppContents, err := aggregateMapper.MapContentPlaceholder(givenMethodeCPH, "tid_test123", "2017-05-15T15:54:32.166Z")
	assert.NoError(t, err, "No error should be thrown for correct mapping.")

	assert.Equal(t, "cdac1f3d-e48c-4618-863c-94bc9d913b9b", actualUppContents[0].GetUUID())
	assert.Equal(t, "cdac1f3d-e48c-4618-863c-94bc9d913b9b", actualUppContents[1].GetUUID())
}

func TestAggregateMapperMappingError_ThrowsError(t *testing.T) {
	mockResolver := new(model.MockIResolver)
	mockValidator := new(model.MockCPHValidator)
	mockContentMapper := new(model.MockCPHMapper)
	mockCompContentMapper := new(model.MockCPHMapper)

	givenMethodeCPH := &model.MethodeContentPlaceholder{}

	mockValidator.On("Validate",
		mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true })).
		Return(nil)

	mockResolver.On("ResolveIdentifier",
		mock.MatchedBy(func(uuid string) bool { return true }),
		mock.MatchedBy(func(tid string) bool { return true }),
		mock.MatchedBy(func(lmd string) bool { return true })).
		Return("abac1f3d-e48c-4618-863c-94bc9d913b9b", nil)

	mockContentMapper.On("MapContentPlaceholder",
		mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }),
		mock.MatchedBy(func(uuid string) bool { return true }),
		mock.MatchedBy(func(tid string) bool { return true }),
		mock.MatchedBy(func(lmd string) bool { return true })).
		Return([]model.UppContent{}, errors.New("Some mapping error"))

	aggregateMapper := NewAggregateCPHMapper(mockResolver, mockValidator, []CPHMapper{mockContentMapper, mockCompContentMapper})

	_, err := aggregateMapper.MapContentPlaceholder(givenMethodeCPH, "tid_test123", "2017-05-15T15:54:32.166Z")
	assert.Error(t, err, "Error should be thrown for error in one of the contained mappers.")
}
