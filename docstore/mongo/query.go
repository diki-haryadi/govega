package mongo

import (
	"fmt"

	"github.com/diki-haryadi/govega/constant"
	"github.com/diki-haryadi/govega/docstore"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func toMongoFilter(q *docstore.QueryOpt) (bson.M, *options.FindOptions) {
	d := bson.M{}
	for _, s := range q.Filter {
		d[s.Field] = toMongoM(s)
	}

	if q.Page > 0 && q.Limit > 0 {
		q.Skip = q.Page * q.Limit
	}

	opt := options.Find()
	opt.SetLimit(int64(q.Limit))
	opt.SetSkip(int64(q.Skip))
	if q.OrderBy != "" {
		dir := -1
		if q.IsAscend {
			dir = 1
		}
		opt.SetSort(bson.D{{Key: q.OrderBy, Value: dir}})
	}

	return d, opt
}

func toMongoM(f docstore.FilterOpt) bson.M {
	switch f.Ops {
	case constant.EQ:
		return bson.M{
			"$eq": f.Value,
		}
	case constant.LT:
		return bson.M{
			"$lt": f.Value,
		}
	case constant.LE:
		return bson.M{
			"$lte": f.Value,
		}
	case constant.GT:
		return bson.M{
			"$gt": f.Value,
		}
	case constant.GE:
		return bson.M{
			"$gte": f.Value,
		}
	case constant.NE:
		return bson.M{
			"$ne": f.Value,
		}
	case constant.IN:
		return bson.M{
			"$in": f.Value,
		}
	case constant.EM:
		return bson.M{
			"$elemMatch": f.Value,
		}
	case constant.RE:
		return bson.M{
			"$regex": primitive.Regex{
				Pattern: fmt.Sprintf("%v", f.Value),
				Options: "i",
			},
		}
	default:
		return nil
	}
}
