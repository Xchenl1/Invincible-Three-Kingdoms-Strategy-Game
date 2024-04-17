package data

import (
	"sgserver/server/game/model"
	"time"
)

type WarReport struct {
	Id                int       `xorm:"id pk autoincr"`
	AttackRid         int       `xorm:"a_rid"`
	DefenseRid        int       `xorm:"d_rid"`
	BegAttackArmy     string    `xorm:"b_a_army"`
	BegDefenseArmy    string    `xorm:"b_d_army"`
	EndAttackArmy     string    `xorm:"e_a_army"`
	EndDefenseArmy    string    `xorm:"e_d_army"`
	BegAttackGeneral  string    `xorm:"b_a_general"`
	BegDefenseGeneral string    `xorm:"b_d_general"`
	EndAttackGeneral  string    `xorm:"e_a_general"`
	EndDefenseGeneral string    `xorm:"e_d_general"`
	Result            int       `xorm:"result"` //0失败，1打平，2胜利
	Rounds            string    `xorm:"rounds"` //回合
	AttackIsRead      bool      `xorm:"a_is_read"`
	DefenseIsRead     bool      `xorm:"d_is_read"`
	DestroyDurable    int       `xorm:"destroy"`
	Occupy            int       `xorm:"occupy"`
	X                 int       `xorm:"x"`
	Y                 int       `xorm:"y"`
	CTime             time.Time `xorm:"ctime"`
}

func (w *WarReport) TableName() string {
	return "war_report"
}

func (w *WarReport) ToModel() interface{} {
	p := model.WarReport{}
	p.CTime = int(w.CTime.UnixNano() / 1e6)
	p.Id = w.Id
	p.AttackRid = w.AttackRid
	p.DefenseRid = w.DefenseRid
	p.BegAttackArmy = w.BegAttackArmy
	p.BegDefenseArmy = w.BegDefenseArmy
	p.EndAttackArmy = w.EndAttackArmy
	p.EndDefenseArmy = w.EndDefenseArmy
	p.BegAttackGeneral = w.BegAttackGeneral
	p.BegDefenseGeneral = w.BegDefenseGeneral
	p.EndAttackGeneral = w.EndAttackGeneral
	p.EndDefenseGeneral = w.EndDefenseGeneral
	p.Result = w.Result
	p.Rounds = w.Rounds
	p.AttackIsRead = w.AttackIsRead
	p.DefenseIsRead = w.DefenseIsRead
	p.DestroyDurable = w.DestroyDurable
	p.Occupy = w.Occupy
	p.X = w.X
	p.X = w.X
	return p
}
