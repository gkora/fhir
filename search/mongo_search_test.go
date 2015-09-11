package search

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	//"github.com/davecgh/go-spew/spew"
	"github.com/intervention-engine/fhir/models"
	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2/dbtest"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MongoSearchSuite struct {
	DBServer      *dbtest.DBServer
	Session       *mgo.Session
	FhirModels    []interface{}
	MongoSearcher *MongoSearcher
	EST           *time.Location
}

var _ = Suite(&MongoSearchSuite{})

func (m *MongoSearchSuite) SetUpSuite(c *C) {
	m.EST = time.FixedZone("EST", -5*60*60)

	//turnOnDebugLog()

	// Set up the database
	m.DBServer = &dbtest.DBServer{}
	m.DBServer.SetPath(c.MkDir())
	//}

	//func (m *MongoSearchSuite) SetUpTest(c *C) {
	m.Session = m.DBServer.Session()
	db := m.Session.DB("fhir-test")
	m.MongoSearcher = &MongoSearcher{db}

	// Read in the data in FHIR format
	data, err := ioutil.ReadFile("../fixtures/search_test_data.json")
	util.CheckErr(err)

	maps := make([]interface{}, 19)
	err = json.Unmarshal(data, &maps)
	util.CheckErr(err)

	for _, resourceMap := range maps {
		r := models.MapToResource(resourceMap, true)
		collection := MongoCollectionNames[reflect.TypeOf(r).Elem().Name()]
		util.CheckErr(db.C(collection).Insert(r))
	}
}

func (m *MongoSearchSuite) TearDownSuite(c *C) {
	//func (m *MongoSearchSuite) TearDownTest(c *C) {
	m.Session.Close()
	m.DBServer.Wipe()
	//}

	//func (m *MongoSearchSuite) TearDownSuite(c *C) {
	m.DBServer.Stop()
}

func turnOnDebugLog() {
	mgo.SetDebug(true)
	var aLogger *log.Logger
	aLogger = log.New(os.Stderr, "", log.LstdFlags)
	mgo.SetLogger(aLogger)
}

// Tests token searches on CodeableConcept

func (m *MongoSearchSuite) TestConditionCodeQueryObjectBySystemAndCode(c *C) {
	q := Query{"Condition", "code=http://snomed.info/sct|123641001"}
	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{
		"code.coding": bson.M{
			"$elemMatch": bson.M{
				"system": bson.RegEx{Pattern: "^http://snomed\\.info/sct$", Options: "i"},
				"code":   bson.RegEx{Pattern: "^123641001$", Options: "i"},
			},
		},
	})
}

func (m *MongoSearchSuite) TestConditionCodeQueryBySystemAndCode(c *C) {
	var conditions []*models.Condition
	q := Query{"Condition", "code=http://snomed.info/sct|123641001"}
	mq := m.MongoSearcher.CreateQuery(q)
	err := mq.All(&conditions)
	util.CheckErr(err)
	c.Assert(conditions, HasLen, 2)
	foundIvd, foundCad := false, false
	for _, cond := range conditions {
		if strings.Contains(cond.Code.Text, "Ischemic Vascular Disease") {
			foundIvd = true
		} else if strings.Contains(cond.Code.Text, "Coronary Artery Disease No MI") {
			foundCad = true
		}
	}
	c.Assert(foundIvd && foundCad, Equals, true)
}

func (m *MongoSearchSuite) TestConditionCodeQueryByWrongCodeSystem(c *C) {
	var conditions []*models.Condition
	q := Query{"Condition", "code=http://hl7.org/fhir/sid/icd-9|123641001"}
	mq := m.MongoSearcher.CreateQuery(q)
	err := mq.All(&conditions)
	util.CheckErr(err)
	c.Assert(conditions, HasLen, 0)
}

func (m *MongoSearchSuite) TestConditionCodeQueryObjectByCode(c *C) {
	q := Query{"Condition", "code=123641001"}

	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{"code.coding.code": bson.RegEx{Pattern: "^123641001$", Options: "i"}})
}

func (m *MongoSearchSuite) TestConditionCodeQueryByCode(c *C) {
	var conditions []*models.Condition
	q := Query{"Condition", "code=123641001"}
	mq := m.MongoSearcher.CreateQuery(q)
	err := mq.All(&conditions)
	util.CheckErr(err)
	c.Assert(conditions, HasLen, 2)
	foundIvd, foundCad := false, false
	for _, cond := range conditions {
		if strings.Contains(cond.Code.Text, "Ischemic Vascular Disease") {
			foundIvd = true
		} else if strings.Contains(cond.Code.Text, "Coronary Artery Disease No MI") {
			foundCad = true
		}
	}
	c.Assert(foundIvd && foundCad, Equals, true)
}

// Tests token searches on Coding

func (m *MongoSearchSuite) TestImagingStudyBodySiteQueryObjectBySystemAndCode(c *C) {
	q := Query{"ImagingStudy", "bodysite=http://snomed.info/sct|67734004"}
	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{
		"series": bson.M{
			"$elemMatch": bson.M{
				"bodySite.system": bson.RegEx{Pattern: "^http://snomed\\.info/sct$", Options: "i"},
				"bodySite.code":   bson.RegEx{Pattern: "^67734004$", Options: "i"},
			},
		},
	})
}

func (m *MongoSearchSuite) TestImagingStudyBodySiteQueryBySystemAndCode(c *C) {
	q := Query{"ImagingStudy", "bodysite=http://snomed.info/sct|67734004"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 1)
}

func (m *MongoSearchSuite) TestImagingStudyBodySiteQueryByWrongCodeSystem(c *C) {
	q := Query{"ImagingStudy", "bodysite=http://hl7.org/fhir/sid/icd-9|67734004"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

// Tests token searches on []Identifier

func (m *MongoSearchSuite) TestEncounterIdentifierQueryObjectBySystemAndValue(c *C) {
	q := Query{"Encounter", "identifier=http://acme.com|1"}
	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{
		"identifier": bson.M{
			"$elemMatch": bson.M{
				"system": bson.RegEx{Pattern: "^http://acme\\.com$", Options: "i"},
				"value":  bson.RegEx{Pattern: "^1$", Options: "i"},
			},
		},
	})
}

func (m *MongoSearchSuite) TestEncounterIdentifierQueryBySystemAndValue(c *C) {
	q := Query{"Encounter", "identifier=http://acme.com|1"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 1)
}

func (m *MongoSearchSuite) TestEncounterIdentifierQueryByWrongSystem(c *C) {
	q := Query{"Encounter", "identifier=http://example.com|1"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

// TODO: Test token searches on boolean, code, string, and ContactPoint

// Tests reference searches by reference id

func (m *MongoSearchSuite) TestConditionReferenceQueryObjectByPatientId(c *C) {
	q := Query{"Condition", "patient=4954037118555241963"}

	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{"patient.referenceid": bson.RegEx{Pattern: "^4954037118555241963$", Options: "i"}})
}

func (m *MongoSearchSuite) TestConditionReferenceQueryByPatientId(c *C) {
	var conditions []*models.Condition

	q := Query{"Condition", "patient=4954037118555241963"}
	mq := m.MongoSearcher.CreateQuery(q)
	err := mq.All(&conditions)
	util.CheckErr(err)
	c.Assert(conditions, HasLen, 5)
}

func (m *MongoSearchSuite) TestConditionReferenceQueryObjectByPatientTypeAndId(c *C) {
	q := Query{"Condition", "patient=Patient/4954037118555241963"}

	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{"patient.referenceid": bson.RegEx{Pattern: "^4954037118555241963$", Options: "i"}, "patient.type": "Patient"})
}

func (m *MongoSearchSuite) TestConditionPatientQueryByTypeAndId(c *C) {
	var conditions []*models.Condition

	q := Query{"Condition", "patient=Patient/4954037118555241963"}
	mq := m.MongoSearcher.CreateQuery(q)
	err := mq.All(&conditions)
	util.CheckErr(err)
	c.Assert(conditions, HasLen, 5)
}

func (m *MongoSearchSuite) TestConditionReferenceQueryObjectByPatientURL(c *C) {
	q := Query{"Condition", "patient=http://acme.com/Patient/123456789"}

	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{"patient.reference": bson.RegEx{Pattern: "^http://acme\\.com/Patient/123456789$", Options: "i"}})
}

// TODO: Test execution of reference search on PatientURL (as above)

// TODO: Test reference searches on chained queries

// Test date searches on DateTime

func (m *MongoSearchSuite) TestConditionOnsetQueryObject(c *C) {
	q := Query{"Condition", "onset=2012-03-01T07:00-05:00"}

	o := m.MongoSearcher.createQueryObject(q)
	// 2012-03-01T07:00-05:00 <= onsetDateTime < 2012-03-01T07:01-05:00
	onsetDt := o["$or"].([]bson.M)[0]["onsetDateTime.time"].(bson.M)
	c.Assert(onsetDt, HasLen, 2)
	gte := onsetDt["$gte"].(time.Time)
	lt := onsetDt["$lt"].(time.Time)
	c.Assert(gte.UnixNano(), Equals, time.Date(2012, time.March, 1, 7, 0, 0, 0, m.EST).UnixNano())
	c.Assert(lt.UnixNano(), Equals, time.Date(2012, time.March, 1, 7, 1, 0, 0, m.EST).UnixNano())
	// 2012-03-01T07:00-05:00 <= onsetPeriod.end && onsetPeriod.start < 2012-03-01T07:01-05:00
	lt = o["$or"].([]bson.M)[1]["onsetPeriod.start.time"].(bson.M)["$lt"].(time.Time)
	gte = o["$or"].([]bson.M)[1]["$or"].([]bson.M)[0]["onsetPeriod.end.time"].(bson.M)["$gte"].(time.Time)
	inf := o["$or"].([]bson.M)[1]["$or"].([]bson.M)[1]["onsetPeriod.end"]
	c.Assert(gte.UnixNano(), Equals, time.Date(2012, time.March, 1, 7, 0, 0, 0, m.EST).UnixNano())
	c.Assert(lt.UnixNano(), Equals, time.Date(2012, time.March, 1, 7, 1, 0, 0, m.EST).UnixNano())
	c.Assert(inf, IsNil)
}

func (m *MongoSearchSuite) TestConditionOnsetQueryToMinute(c *C) {
	var conditions []*models.Condition
	q := Query{"Condition", "onset=2012-03-01T07:00-05:00"}
	mq := m.MongoSearcher.CreateQuery(q)
	err := mq.All(&conditions)
	util.CheckErr(err)
	c.Assert(conditions, HasLen, 1)

	cond := conditions[0]
	cond2 := &models.Condition{}
	err = m.Session.DB("fhir-test").C("conditions").FindId("8664777288161060797").One(cond2)

	c.Assert(cond, DeepEquals, cond2)
}

func (m *MongoSearchSuite) TestConditionOnsetQueryToDay(c *C) {
	q := Query{"Condition", "onset=2012-03-01"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 5)
}

func (m *MongoSearchSuite) TestConditionOnsetQueryWrongTime(c *C) {
	q := Query{"Condition", "onset=2012-03-01T08:00-05:00"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

// Test date searches on Period

func (m *MongoSearchSuite) TestEncounterPeriodQueryObject(c *C) {
	q := Query{"Encounter", "date=2012-11-01T08:50:00-05:00"}

	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, HasLen, 2)
	// 2012-11-01T08:50:00-05:00 <= onsetPeriod.end && onsetPeriod.start < 2012-11-01T08:50:00-05:00
	lt := o["period.start.time"].(bson.M)["$lt"].(time.Time)
	gte := o["$or"].([]bson.M)[0]["period.end.time"].(bson.M)["$gte"].(time.Time)
	inf := o["$or"].([]bson.M)[1]["period.end"]
	c.Assert(gte.UnixNano(), Equals, time.Date(2012, time.November, 1, 8, 50, 0, 0, m.EST).UnixNano())
	c.Assert(lt.UnixNano(), Equals, time.Date(2012, time.November, 1, 8, 50, 1, 0, m.EST).UnixNano())
	c.Assert(inf, IsNil)
}

func (m *MongoSearchSuite) TestEncounterPeriodQuery(c *C) {
	q := Query{"Encounter", "date=2012-11-01T08:50:00-05:00"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 2)

	q = Query{"Encounter", "date=2012-11-01T08:30:00-05:00"}
	mq = m.MongoSearcher.CreateQuery(q)
	num, err = mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 1)
}

func (m *MongoSearchSuite) TestEncounterPeriodQueryWrongTime(c *C) {
	q := Query{"Encounter", "date=2012-11-01T07:50:00-05:00"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

// TODO: Test date searches on date, instant, and Timing

// Test number searches on positiveInt

func (m *MongoSearchSuite) TestImmunizationDoseSequenceNumberQueryObject(c *C) {
	q := Query{"Immunization", "dose-sequence=1"}

	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{
		"vaccinationProtocol": bson.M{
			"$elemMatch": bson.M{
				"doseSequence": bson.M{
					"$gte": float64(0.5),
					"$lt":  float64(1.5),
				},
			},
		},
	})
}

func (m *MongoSearchSuite) TestImmunizationDoseSequenceNumberQuery(c *C) {
	q := Query{"Immunization", "dose-sequence=1"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 1)
}

func (m *MongoSearchSuite) TestImmunizationDoseSequenceWrongNumberQuery(c *C) {
	q := Query{"Immunization", "dose-sequence=0"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)

	q = Query{"Immunization", "dose-sequence=2"}
	mq = m.MongoSearcher.CreateQuery(q)
	num, err = mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

// TODO: Test number searches on decimal, integer, and unsignedInt

// Test string searches on string

func (m *MongoSearchSuite) TestDeviceStringQueryObject(c *C) {
	q := Query{"Device", "manufacturer=Acme"}

	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{"manufacturer": bson.RegEx{Pattern: "^Acme", Options: "i"}})
}

func (m *MongoSearchSuite) TestDeviceStringQuery(c *C) {
	q := Query{"Device", "manufacturer=Acme"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 1)

	dev := &models.Device{}
	err = mq.One(dev)
	util.CheckErr(err)

	dev2 := &models.Device{}
	err = m.Session.DB("fhir-test").C("devices").FindId("7045606679745526995").One(dev2)

	c.Assert(dev, DeepEquals, dev2)
}

func (m *MongoSearchSuite) TestNonMatchingDeviceStringQuery(c *C) {
	q := Query{"Device", "manufacturer=Zinc"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

// Test string searches on HumanName

func (m *MongoSearchSuite) TestPatientNameStringQueryObject(c *C) {
	q := Query{"Patient", "name=Peters"}

	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{
		"$or": []bson.M{
			bson.M{"name.text": bson.RegEx{Pattern: "^Peters", Options: "i"}},
			bson.M{"name.family": bson.RegEx{Pattern: "^Peters", Options: "i"}},
			bson.M{"name.given": bson.RegEx{Pattern: "^Peters", Options: "i"}},
		},
	})
}

func (m *MongoSearchSuite) TestPatientNameStringQuery(c *C) {
	q := Query{"Patient", "name=Peters"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 1)
}

func (m *MongoSearchSuite) TestNonMatchingPatientNameStringQuery(c *C) {
	q := Query{"Patient", "name=Peterson"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

// Test string searches on Address

func (m *MongoSearchSuite) TestPatientAddressStringQueryObject(c *C) {
	q := Query{"Patient", "address=AK"}

	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{
		"$or": []bson.M{
			bson.M{"address.text": bson.RegEx{Pattern: "^AK", Options: "i"}},
			bson.M{"address.line": bson.RegEx{Pattern: "^AK", Options: "i"}},
			bson.M{"address.city": bson.RegEx{Pattern: "^AK", Options: "i"}},
			bson.M{"address.state": bson.RegEx{Pattern: "^AK", Options: "i"}},
			bson.M{"address.postalCode": bson.RegEx{Pattern: "^AK", Options: "i"}},
			bson.M{"address.country": bson.RegEx{Pattern: "^AK", Options: "i"}},
		},
	})
}

func (m *MongoSearchSuite) TestPatientAddressStringQuery(c *C) {
	q := Query{"Patient", "address=AK"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 1)
}

func (m *MongoSearchSuite) TestNonMatchingPatientAddressStringQuery(c *C) {
	q := Query{"Patient", "address=CA"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

// Test quantity searches on Quantity

func (m *MongoSearchSuite) TestValueQuantityQueryObjectByValueAndUnit(c *C) {
	q := Query{"Observation", "value-quantity=185||lbs"}
	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{
		"valueQuantity.value": bson.M{
			"$gte": float64(184.5),
			"$lt":  float64(185.5),
		},
		"$or": []bson.M{
			bson.M{"valueQuantity.code": bson.RegEx{Pattern: "^lbs$", Options: "i"}},
			bson.M{"valueQuantity.units": bson.RegEx{Pattern: "^lbs$", Options: "i"}},
		},
	})
}

func (m *MongoSearchSuite) TestValueQuantityQueryByValueAndUnit(c *C) {
	q := Query{"Observation", "value-quantity=185||lbs"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 1)
}

func (m *MongoSearchSuite) TestValueQuantityQueryByValueAndCode(c *C) {
	q := Query{"Observation", "value-quantity=185||[lb_av]"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 1)
}

func (m *MongoSearchSuite) TestValueQuantityQueryByWrongValueAndUnit(c *C) {
	q := Query{"Observation", "value-quantity=186||lbs"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

func (m *MongoSearchSuite) TestValueQuantityQueryByValueAndWrongUnit(c *C) {
	q := Query{"Observation", "value-quantity=185||pounds"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

func (m *MongoSearchSuite) TestValueQuantityQueryObjectByValueAndSystemAndCode(c *C) {
	q := Query{"Observation", "value-quantity=185|http://unitsofmeasure.org|[lb_av]"}
	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{
		"valueQuantity.value": bson.M{
			"$gte": float64(184.5),
			"$lt":  float64(185.5),
		},
		"valueQuantity.code":   bson.RegEx{Pattern: "^\\[lb_av\\]$", Options: "i"},
		"valueQuantity.system": bson.RegEx{Pattern: "^http://unitsofmeasure\\.org$", Options: "i"},
	})
}

func (m *MongoSearchSuite) TestValueQuantityQueryByValueAndSystemAndCode(c *C) {
	q := Query{"Observation", "value-quantity=185|http://unitsofmeasure.org|[lb_av]"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 1)
}

func (m *MongoSearchSuite) TestValueQuantityQueryByWrongValueAndSystemAndCode(c *C) {
	q := Query{"Observation", "value-quantity=184|http://unitsofmeasure.org|[lb_av]"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

func (m *MongoSearchSuite) TestValueQuantityQueryByValueAndWrongSystemAndCode(c *C) {
	q := Query{"Observation", "value-quantity=185|http://loinc.org|[lb_av]"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

func (m *MongoSearchSuite) TestValueQuantityQueryByValueAndSystemAndWrongCode(c *C) {
	q := Query{"Observation", "value-quantity=185|http://unitsofmeasure.org|lbs"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

// TODO: Test quantity searches on Money, SimpleQuantity, Duration, Count, Distance, and Age

// Test URI searches on URI

func (m *MongoSearchSuite) TestSubscriptionURLQueryObject(c *C) {
	q := Query{"Subscription", "url=https://biliwatch.com/customers/mount-auburn-miu/on-result"}
	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{
		"channel.endpoint": "https://biliwatch.com/customers/mount-auburn-miu/on-result",
	})
}

func (m *MongoSearchSuite) TestSubscriptionURLQuery(c *C) {
	q := Query{"Subscription", "url=https://biliwatch.com/customers/mount-auburn-miu/on-result"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 1)
}

// TODO: Test composite searches

// Tests special searches on _id

func (m *MongoSearchSuite) TestConditionIdQueryObject(c *C) {
	q := Query{"Condition", "_id=123456789"}

	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{"_id": bson.RegEx{Pattern: "^123456789$", Options: "i"}})
}

func (m *MongoSearchSuite) TestConditionIdQuery(c *C) {
	q := Query{"Condition", "_id=8664777288161060797"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 1)

	cond := &models.Condition{}
	err = mq.One(cond)
	util.CheckErr(err)

	cond2 := &models.Condition{}
	err = m.Session.DB("fhir-test").C("conditions").FindId("8664777288161060797").One(cond2)

	c.Assert(cond, DeepEquals, cond2)
}

// TODO: Test special searches: _content, _lastUpdated, _profile, _query, _security, _tag, _text

// Test searches with multiple values
func (m *MongoSearchSuite) TestConditionMultipleCodesQueryObject(c *C) {
	q := Query{"Condition", "code=http://hl7.org/fhir/sid/icd-9|428.0,http://snomed.info/sct|981000124106,http://hl7.org/fhir/sid/icd-10|I20.0"}

	o := m.MongoSearcher.createQueryObject(q)
	c.Assert(o, DeepEquals, bson.M{
		"$or": []bson.M{
			bson.M{
				"code.coding": bson.M{
					"$elemMatch": bson.M{
						"system": bson.RegEx{Pattern: "^http://hl7\\.org/fhir/sid/icd-9$", Options: "i"},
						"code":   bson.RegEx{Pattern: "^428\\.0$", Options: "i"},
					}},
			},
			bson.M{
				"code.coding": bson.M{
					"$elemMatch": bson.M{
						"system": bson.RegEx{Pattern: "^http://snomed\\.info/sct$", Options: "i"},
						"code":   bson.RegEx{Pattern: "^981000124106$", Options: "i"},
					}},
			},
			bson.M{
				"code.coding": bson.M{
					"$elemMatch": bson.M{
						"system": bson.RegEx{Pattern: "^http://hl7\\.org/fhir/sid/icd-10$", Options: "i"},
						"code":   bson.RegEx{Pattern: "^I20\\.0$", Options: "i"},
					}},
			},
		},
	})
}

func (m *MongoSearchSuite) TestConditionMultipleCodesQuery(c *C) {
	q := Query{"Condition", "code=http://hl7.org/fhir/sid/icd-9|428.0,http://snomed.info/sct|981000124106,http://hl7.org/fhir/sid/icd-10|I20.0"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 4)
}

func (m *MongoSearchSuite) TestConditionMultipleCodesWrongICD10Query(c *C) {
	q := Query{"Condition", "code=http://hl7.org/fhir/sid/icd-9|428.0,http://snomed.info/sct|981000124106,http://hl7.org/fhir/sid/icd-10|I21.0"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 2)
}

// Test searches with multiple parameters
func (m *MongoSearchSuite) TestConditionPatientAndCodeAndOnsetQueryObject(c *C) {
	q := Query{"Condition", "patient=4954037118555241963&code=http://hl7.org/fhir/sid/icd-9|428.0&onset=2012-03-01T07:00-05:00"}

	o := m.MongoSearcher.createQueryObject(q)
	// Make sure only the expected elements are there
	c.Assert(o, HasLen, 3)

	// Check the patient part of the query
	c.Assert(o["patient.referenceid"], DeepEquals, bson.RegEx{Pattern: "^4954037118555241963$", Options: "i"})

	// Check the code part of the query
	c.Assert(o["code.coding"], DeepEquals, bson.M{
		"$elemMatch": bson.M{
			"system": bson.RegEx{Pattern: "^http://hl7\\.org/fhir/sid/icd-9$", Options: "i"},
			"code":   bson.RegEx{Pattern: "^428\\.0$", Options: "i"},
		},
	})

	// Check the onset part of the query
	// 2012-03-01T07:00-05:00 <= onsetDateTime < 2012-03-01T07:01-05:00
	onsetDt := o["$or"].([]bson.M)[0]["onsetDateTime.time"].(bson.M)
	c.Assert(onsetDt, HasLen, 2)
	gte := onsetDt["$gte"].(time.Time)
	lt := onsetDt["$lt"].(time.Time)
	c.Assert(gte.UnixNano(), Equals, time.Date(2012, time.March, 1, 7, 0, 0, 0, m.EST).UnixNano())
	c.Assert(lt.UnixNano(), Equals, time.Date(2012, time.March, 1, 7, 1, 0, 0, m.EST).UnixNano())
	// 2012-03-01T07:00-05:00 <= onsetPeriod.end && onsetPeriod.start < 2012-03-01T07:01-05:00
	lt = o["$or"].([]bson.M)[1]["onsetPeriod.start.time"].(bson.M)["$lt"].(time.Time)
	gte = o["$or"].([]bson.M)[1]["$or"].([]bson.M)[0]["onsetPeriod.end.time"].(bson.M)["$gte"].(time.Time)
	inf := o["$or"].([]bson.M)[1]["$or"].([]bson.M)[1]["onsetPeriod.end"]
	c.Assert(gte.UnixNano(), Equals, time.Date(2012, time.March, 1, 7, 0, 0, 0, m.EST).UnixNano())
	c.Assert(lt.UnixNano(), Equals, time.Date(2012, time.March, 1, 7, 1, 0, 0, m.EST).UnixNano())
	c.Assert(inf, IsNil)
}

func (m *MongoSearchSuite) TestConditionPatientAndCodeAndOnsetQuery(c *C) {
	q := Query{"Condition", "patient=4954037118555241963&code=http://hl7.org/fhir/sid/icd-9|428.0&onset=2012-03-01T07:00-05:00"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 1)
}

func (m *MongoSearchSuite) TestConditionWrongPatientAndCodeAndOnsetQuery(c *C) {
	q := Query{"Condition", "patient=123456789&code=http://hl7.org/fhir/sid/icd-9|428.0&onset=2012-03-01T07:00-05:00"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

func (m *MongoSearchSuite) TestConditionPatientAndWrongCodeAndOnsetQuery(c *C) {
	q := Query{"Condition", "patient=4954037118555241963&code=http://snomed.info/sct|981000124106&onset=2012-03-01T07:00-05:00"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

func (m *MongoSearchSuite) TestConditionPatientAndCodeAndWrongOnsetQuery(c *C) {
	q := Query{"Condition", "patient=4954037118555241963&code=http://hl7.org/fhir/sid/icd-9|428.0&onset=2012-03-01T07:05-05:00"}
	mq := m.MongoSearcher.CreateQuery(q)
	num, err := mq.Count()
	util.CheckErr(err)
	c.Assert(num, Equals, 0)
}

// Test multiple parameters with multiple values
func (m *MongoSearchSuite) TestConditionPatientAndMultipleCodesQueryObject(c *C) {
	q := Query{"Condition", "patient=4954037118555241963&code=http://hl7.org/fhir/sid/icd-9|428.0,http://snomed.info/sct|981000124106"}

	o := m.MongoSearcher.createQueryObject(q)
	// Make sure only the expected elements are there
	c.Assert(o, HasLen, 2)

	// Check the patient part of the query
	c.Assert(o["patient.referenceid"], DeepEquals, bson.RegEx{Pattern: "^4954037118555241963$", Options: "i"})

	// Check the code part of the query
	c.Assert(o["$or"], DeepEquals, []bson.M{
		bson.M{
			"code.coding": bson.M{
				"$elemMatch": bson.M{
					"system": bson.RegEx{Pattern: "^http://hl7\\.org/fhir/sid/icd-9$", Options: "i"},
					"code":   bson.RegEx{Pattern: "^428\\.0$", Options: "i"},
				},
			},
		},
		bson.M{
			"code.coding": bson.M{
				"$elemMatch": bson.M{
					"system": bson.RegEx{Pattern: "^http://snomed\\.info/sct$", Options: "i"},
					"code":   bson.RegEx{Pattern: "^981000124106$", Options: "i"},
				},
			},
		},
	})
}

func (m *MongoSearchSuite) TestConditionMultiplePatientAndMultipleCodesQueryObject(c *C) {
	q := Query{"Condition", "patient=4954037118555241963,123456789,ABCDEFG&code=http://hl7.org/fhir/sid/icd-9|428.0,http://snomed.info/sct|981000124106"}

	o := m.MongoSearcher.createQueryObject(q)
	// Make sure only the expected elements are there
	c.Assert(o, HasLen, 2)

	expectedTopOr := []bson.M{
		{"patient.referenceid": bson.RegEx{Pattern: "^4954037118555241963$", Options: "i"}},
		{"patient.referenceid": bson.RegEx{Pattern: "^123456789$", Options: "i"}},
		{"patient.referenceid": bson.RegEx{Pattern: "^ABCDEFG$", Options: "i"}},
	}

	expectedNestedOr := []bson.M{
		bson.M{
			"code.coding": bson.M{
				"$elemMatch": bson.M{
					"system": bson.RegEx{Pattern: "^http://hl7\\.org/fhir/sid/icd-9$", Options: "i"},
					"code":   bson.RegEx{Pattern: "^428\\.0$", Options: "i"},
				},
			},
		},
		bson.M{
			"code.coding": bson.M{
				"$elemMatch": bson.M{
					"system": bson.RegEx{Pattern: "^http://snomed\\.info/sct$", Options: "i"},
					"code":   bson.RegEx{Pattern: "^981000124106$", Options: "i"},
				},
			},
		},
	}

	// Can't guarantee what order they're processed in, so need to account for that in the test
	if len(o["$or"].([]bson.M)) == 2 {
		expectedTopOr, expectedNestedOr = expectedNestedOr, expectedTopOr
	}

	c.Assert(o["$or"], DeepEquals, expectedTopOr)
	c.Assert(o["$and"], DeepEquals, []bson.M{
		bson.M{"$or": expectedNestedOr},
	})
}

// Test that invalid search parameters PANIC (to ensure people know they are broken)
func (m *MongoSearchSuite) TestInvalidSearchParameterPanics(c *C) {
	q := Query{"Condition", "abatement=2012"}
	c.Assert(func() { m.MongoSearcher.CreateQuery(q) }, Panics, InvalidSearchError("Condition does not support search parameter: abatement"))
}

// Test that unimplemented features PANIC (to ensure people know they are broken)
func (m *MongoSearchSuite) TestCompositeSearchPanics(c *C) {
	q := Query{"Group", "characteristic-value=gender$male"}
	c.Assert(func() { m.MongoSearcher.CreateQuery(q) }, Panics, UnsupportedError("composite search parameters"))
}

func (m *MongoSearchSuite) TestPrefixedDateSearchPanics(c *C) {
	q := Query{"Condition", "onset=lt2012"}
	c.Assert(func() { m.MongoSearcher.CreateQuery(q) }, Panics, UnsupportedError("date search prefix: lt"))
}

func (m *MongoSearchSuite) TestPrefixedNumberSearchPanics(c *C) {
	q := Query{"Immunization", "dose-sequence=gt1"}
	c.Assert(func() { m.MongoSearcher.CreateQuery(q) }, Panics, UnsupportedError("number search prefix: gt"))
}

func (m *MongoSearchSuite) TestPrefixedQuantitySearchPanics(c *C) {
	q := Query{"Observation", "value-quantity=ap1||mg"}
	c.Assert(func() { m.MongoSearcher.CreateQuery(q) }, Panics, UnsupportedError("quantity search prefix: ap"))
}

func (m *MongoSearchSuite) TestModifierSearchPanics(c *C) {
	q := Query{"Condition", "code:text=headache"}
	c.Assert(func() { m.MongoSearcher.CreateQuery(q) }, Panics, UnsupportedError("search modifier: :text"))
}

func (m *MongoSearchSuite) TestChainedSearchPanics(c *C) {
	q := Query{"Condition", "patient.gender=male"}
	c.Assert(func() { m.MongoSearcher.CreateQuery(q) }, Panics, UnsupportedError("chained search parameters"))
}

func (m *MongoSearchSuite) TestSpecialSearchParameterPanics(c *C) {
	q := Query{"Condition", "onset=2012&_sort:asc=onset"}
	c.Assert(func() { m.MongoSearcher.CreateQuery(q) }, Panics, UnsupportedError("special search parameter: _sort:asc"))
}

// Test internally used functions

func (m *MongoSearchSuite) TestBuildBsonForCompositeCriteriaAndPathWithArrayAncestor(c *C) {
	b := buildBSON("a.[]b.c.d.e", bson.M{"x": 1, "y": 2})
	c.Assert(b, DeepEquals, bson.M{
		"a.b": bson.M{
			"$elemMatch": bson.M{
				"c.d.e.x": 1,
				"c.d.e.y": 2,
			},
		},
	})
}

func (m *MongoSearchSuite) TestBuildBsonForCompositeCriteriaAndPathWithArrayAncestors(c *C) {
	b := buildBSON("a.[]b.[]c.d.e", bson.M{"x": 1, "y": 2})
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c": bson.M{
			"$elemMatch": bson.M{
				"d.e.x": 1,
				"d.e.y": 2,
			},
		},
	})
}

func (m *MongoSearchSuite) TestBuildBsonForCompositeCriteriaAndPathWithArrayAncestorsAndLeaf(c *C) {
	b := buildBSON("a.[]b.[]c.d.[]e", bson.M{"x": 1, "y": 2})
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c.d.e": bson.M{
			"$elemMatch": bson.M{
				"x": 1,
				"y": 2,
			},
		},
	})
}

func (m *MongoSearchSuite) TestBuildBsonForCompositeCriteriaAndPathWithArrayLeaf(c *C) {
	b := buildBSON("a.b.c.d.[]e", bson.M{"x": 1, "y": 2})
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c.d.e": bson.M{
			"$elemMatch": bson.M{
				"x": 1,
				"y": 2,
			},
		},
	})
}

func (m *MongoSearchSuite) TestBuildBsonForSingleKeyCriteriaAndPathWithArrayAncestor(c *C) {
	b := buildBSON("a.[]b.c.d.e", bson.M{"x": 1})
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c.d.e.x": 1,
	})
}

func (m *MongoSearchSuite) TestBuildBsonForSingleKeyCriteriaAndPathWithArrayLeaf(c *C) {
	b := buildBSON("a.b.c.d.[]e", bson.M{"x": 1})
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c.d.e.x": 1,
	})
}

func (m *MongoSearchSuite) TestBuildBsonForSingularCriteriaAndPathWithArrayAncestor(c *C) {
	b := buildBSON("a.[]b.c.d.e", 1)
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c.d.e": 1,
	})
}

func (m *MongoSearchSuite) TestBuildBsonForSingularCriteriaAndPathWithArrayLeaf(c *C) {
	b := buildBSON("a.b.c.d.[]e", 1)
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c.d.e": 1,
	})
}

func (m *MongoSearchSuite) TestBuildBsonForCompositeCriteriaAndPathWithoutArrays(c *C) {
	b := buildBSON("a.b.c.d.e", bson.M{"x": 1, "y": 2})
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c.d.e.x": 1,
		"a.b.c.d.e.y": 2,
	})
}

func (m *MongoSearchSuite) TestBuildBsonForSingleKeyCriteriaAndPathWithoutArrays(c *C) {
	b := buildBSON("a.b.c.d.e", bson.M{"x": 1})
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c.d.e.x": 1,
	})
}

func (m *MongoSearchSuite) TestBuildBsonForSingularCriteriaAndPathWithoutArrays(c *C) {
	b := buildBSON("a.b.c.d.e", 1)
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c.d.e": 1,
	})
}

func (m *MongoSearchSuite) TestBuildBsonWithSingleOrCriteria(c *C) {
	b := buildBSON("a.b.c.d.e", bson.M{"$or": []bson.M{bson.M{"x": 1}, bson.M{"y": 2}}})
	c.Assert(b, DeepEquals, bson.M{
		"$or": []bson.M{
			bson.M{
				"a.b.c.d.e.x": 1,
			},
			bson.M{
				"a.b.c.d.e.y": 2,
			},
		},
	})
}

func (m *MongoSearchSuite) TestBuildBsonWithAndOrCriteria(c *C) {
	b := buildBSON("a.b.c.d.e", bson.M{"z": 3, "$or": []bson.M{bson.M{"x": 1}, bson.M{"y": 2}}})
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c.d.e.z": 3,
		"$or": []bson.M{
			bson.M{
				"a.b.c.d.e.x": 1,
			},
			bson.M{
				"a.b.c.d.e.y": 2,
			},
		},
	})
}

func (m *MongoSearchSuite) TestBuildBsonWithAndComplexOrCriteria(c *C) {
	b := buildBSON("a.b.c.d.e", bson.M{"z": 3, "$or": []bson.M{bson.M{"x": 1, "a": 99}, bson.M{"y": 2}}})
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c.d.e.z": 3,
		"$or": []bson.M{
			bson.M{
				"a.b.c.d.e.x": 1,
				"a.b.c.d.e.a": 99,
			},
			bson.M{
				"a.b.c.d.e.y": 2,
			},
		},
	})
}

func (m *MongoSearchSuite) TestBuildBsonWithSingleOrCriteriaAndArrayAncestors(c *C) {
	b := buildBSON("a.b.[]c.d.e", bson.M{"$or": []bson.M{bson.M{"x": 1}, bson.M{"y": 2}}})
	c.Assert(b, DeepEquals, bson.M{
		"$or": []bson.M{
			bson.M{
				"a.b.c.d.e.x": 1,
			},
			bson.M{
				"a.b.c.d.e.y": 2,
			},
		},
	})
}

func (m *MongoSearchSuite) TestBuildBsonWithSingleOrCriteriaAndArrayLeaf(c *C) {
	b := buildBSON("a.b.c.d.[]e", bson.M{"$or": []bson.M{bson.M{"x": 1}, bson.M{"y": 2}}})
	c.Assert(b, DeepEquals, bson.M{
		"$or": []bson.M{
			bson.M{
				"a.b.c.d.e.x": 1,
			},
			bson.M{
				"a.b.c.d.e.y": 2,
			},
		},
	})
}

func (m *MongoSearchSuite) TestBuildBsonWithAndOrCriteriaAndArrayAncenstors(c *C) {
	b := buildBSON("a.b.[]c.d.e", bson.M{"z": 3, "$or": []bson.M{bson.M{"x": 1, "a": 99}, bson.M{"y": 2}}})
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c": bson.M{
			"$elemMatch": bson.M{
				"d.e.z": 3,
				"$or": []bson.M{
					bson.M{
						"d.e.x": 1,
						"d.e.a": 99,
					},
					bson.M{
						"d.e.y": 2,
					},
				},
			},
		},
	})
}

func (m *MongoSearchSuite) TestBuildBsonWithAndOrCriteriaAndArrayLeaf(c *C) {
	b := buildBSON("a.b.c.d.[]e", bson.M{"z": 3, "$or": []bson.M{bson.M{"x": 1, "a": 99}, bson.M{"y": 2}}})
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c.d.e": bson.M{
			"$elemMatch": bson.M{
				"z": 3,
				"$or": []bson.M{
					bson.M{
						"x": 1,
						"a": 99,
					},
					bson.M{
						"y": 2,
					},
				},
			},
		},
	})
}

func (m *MongoSearchSuite) TestBuildBsonWithQueryOperators(c *C) {
	b := buildBSON("a.b.c.d.e", bson.M{"$gte": 0, "$lt": 1})
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c.d.e": bson.M{
			"$gte": 0,
			"$lt":  1,
		},
	})
}

func (m *MongoSearchSuite) TestBuildBsonWithQueryOperatorsAndArrayAncestors(c *C) {
	b := buildBSON("a.b.[]c.d.e", bson.M{"$gte": 0, "$lt": 1})
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c": bson.M{
			"$elemMatch": bson.M{
				"d.e": bson.M{
					"$gte": 0,
					"$lt":  1,
				},
			},
		},
	})
}

func (m *MongoSearchSuite) TestBuildBsonWithQueryOperatorsAndArrayLeaf(c *C) {
	b := buildBSON("a.b.c.d.[]e", bson.M{"$gte": 0, "$lt": 1})
	c.Assert(b, DeepEquals, bson.M{
		"a.b.c.d.e": bson.M{
			"$elemMatch": bson.M{
				"$gte": 0,
				"$lt":  1,
			},
		},
	})
}

func (m *MongoSearchSuite) TestMergeObjectsWithNoCommonKeys(c *C) {
	o1 := bson.M{
		"a.b.c": 1,
		"d.e.f": 2,
	}
	o2 := bson.M{
		"g.h.i": 3,
		"j.k.l": 4,
	}
	merge(o1, o2)

	c.Assert(o1, DeepEquals, bson.M{
		"a.b.c": 1,
		"d.e.f": 2,
		"g.h.i": 3,
		"j.k.l": 4,
	})
}

func (m *MongoSearchSuite) TestMergeObjectsWithCommonKeys(c *C) {
	o1 := bson.M{
		"a.b.c": 1,
		"d.e.f": 2,
	}
	o2 := bson.M{
		"g.h.i": 3,
		"a.b.c": 4,
	}
	merge(o1, o2)

	c.Assert(o1, DeepEquals, bson.M{
		"a.b.c": 1,
		"d.e.f": 2,
		"g.h.i": 3,
		"$and": []bson.M{
			bson.M{"a.b.c": 4},
		},
	})
}

func (m *MongoSearchSuite) TestMergeObjectsWithExistingAndOnLeft(c *C) {
	o1 := bson.M{
		"a.b.c": 1,
		"d.e.f": 2,
		"$and": []bson.M{
			bson.M{"a.b.c": 5},
		},
	}
	o2 := bson.M{
		"g.h.i": 3,
		"a.b.c": 4,
	}
	merge(o1, o2)

	c.Assert(o1, DeepEquals, bson.M{
		"a.b.c": 1,
		"d.e.f": 2,
		"g.h.i": 3,
		"$and": []bson.M{
			bson.M{"a.b.c": 5},
			bson.M{"a.b.c": 4},
		},
	})
}

func (m *MongoSearchSuite) TestMergeObjectsWithExistingAndOnRight(c *C) {
	o1 := bson.M{
		"a.b.c": 1,
		"d.e.f": 2,
	}
	o2 := bson.M{
		"g.h.i": 3,
		"a.b.c": 4,
		"$and": []bson.M{
			bson.M{"a.b.c": 5},
		},
	}
	merge(o1, o2)

	// Can't guarantee order of $and, so check one-by-one
	c.Assert(o1, HasLen, 4)
	c.Assert(o1["a.b.c"], Equals, 1)
	c.Assert(o1["d.e.f"], Equals, 2)
	c.Assert(o1["g.h.i"], Equals, 3)
	c.Assert(o1["$and"], FitsTypeOf, []bson.M{})
	found4, found5 := false, false
	for _, m := range o1["$and"].([]bson.M) {
		c.Assert(m, HasLen, 1)
		switch m["a.b.c"] {
		case 4:
			found4 = true
		case 5:
			found5 = true
		}
	}
	c.Assert(found4 && found5, Equals, true)
}

func (m *MongoSearchSuite) TestMergeObjectsWithExistingAndOnLeftAndRight(c *C) {
	o1 := bson.M{
		"a.b.c": 1,
		"d.e.f": 2,
		"$and": []bson.M{
			bson.M{"a.b.c": 6},
		},
	}
	o2 := bson.M{
		"g.h.i": 3,
		"a.b.c": 4,
		"$and": []bson.M{
			bson.M{"a.b.c": 5},
		},
	}
	merge(o1, o2)

	// Can't guarantee order of $and, so check one-by-one
	c.Assert(o1, HasLen, 4)
	c.Assert(o1["a.b.c"], Equals, 1)
	c.Assert(o1["d.e.f"], Equals, 2)
	c.Assert(o1["g.h.i"], Equals, 3)
	c.Assert(o1["$and"], FitsTypeOf, []bson.M{})
	found4, found5, found6 := false, false, false
	for _, m := range o1["$and"].([]bson.M) {
		c.Assert(m, HasLen, 1)
		switch m["a.b.c"] {
		case 4:
			found4 = true
		case 5:
			found5 = true
		case 6:
			found6 = true
		}
	}
	c.Assert(found4 && found5 && found6, Equals, true)
}