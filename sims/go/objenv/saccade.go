// Copyright (c) 2020, The CCNLab Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"math"

	"github.com/emer/emergent/env"
	"github.com/emer/emergent/evec"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/minmax"
	"github.com/goki/gi/mat32"
)

// Saccade implements saccading logic for generating visual saccades
// around a 2D World plane, with a moving object that must remain
// in view.  Generates the track of the object.
// World size is defined as -1..1 in normalized units.
type Saccade struct {
	TrajLenRange minmax.Int `desc:"range of trajectory lengths (time steps)"`
	FixDurRange  minmax.Int `desc:"range of fixation durations"`
	SacGenMax    float32    `desc:"maximum saccade size"`
	VelGenMax    float32    `desc:"maximum object velocity"`
	ZeroVelP     float32    `desc:"probability of zero velocity object motion as a discrete option prior to computing random velocity"`
	Margin       float32    `desc:"edge around World to not look past"`
	ViewPct      float32    `desc:"size of view as proportion of -1..1 world size"`
	WorldVisSz   evec.Vec2i `desc:"visualization size of world -- for debug visualization"`
	ViewVisSz    evec.Vec2i `desc:"visualization size of view -- for debug visualization"`
	AddRows      bool       `desc:"add rows to Table for each step (for debugging) -- else re-use 0"`

	// State below here
	Table       *etable.Table    `desc:"table showing visualization of state"`
	WorldTsr    *etensor.Float32 `desc:"tensor state showing world position of obj"`
	ViewTsr     *etensor.Float32 `desc:"tensor state showing view position of obj"`
	TrajLen     int              `desc:"current trajectory length"`
	FixDur      int              `desc:"current fixation duration"`
	Tick        env.Ctr          `desc:"tick counter within trajectory"`
	SacTick     env.Ctr          `desc:"tick counter within current fixation"`
	World       minmax.F32       `desc:"World minus margin"`
	View        minmax.F32       `desc:"View minus margin"`
	ObjPos      mat32.Vec2       `desc:"object position, in world coordinates"`
	ObjViewPos  mat32.Vec2       `desc:"object position, in view coordinates"`
	ObjVel      mat32.Vec2       `desc:"object velocity, in world coordinates"`
	ObjPosNext  mat32.Vec2       `desc:"next object position, in world coordinates"`
	ObjVelNext  mat32.Vec2       `desc:"next object velocity, in world coordinates"`
	EyePos      mat32.Vec2       `desc:"eye position, in world coordinates"`
	SacPlan     mat32.Vec2       `desc:"eye movement plan, in world coordinates"`
	Saccade     mat32.Vec2       `desc:"current trial eye movement, in world coordinates"`
	NewTraj     bool             `desc:"true if new trajectory started"`
	NewTrajNext bool             `desc:"true if next trial will be a new trajectory"`
	DidSaccade  bool             `desc:"did saccade on current trial"`
}

// Defaults sets generic defaults -- use ParamSet to override
func (sc *Saccade) Defaults() {
	sc.TrajLenRange.Set(8, 8)
	sc.FixDurRange.Set(2, 2)
	sc.SacGenMax = 0.4
	sc.VelGenMax = 0.4
	sc.ZeroVelP = 0
	sc.Margin = 0.1
	sc.ViewPct = 0.5
	sc.WorldVisSz.Set(24, 24)
	sc.ViewVisSz.Set(16, 16)
	sc.AddRows = true
}

func (sc *Saccade) Init() {
	sc.World.Max = 1 - sc.Margin
	sc.World.Min = -1 + sc.Margin
	sc.View.Max = sc.ViewPct - sc.Margin
	sc.View.Min = -sc.ViewPct + sc.Margin
	if sc.Table == nil {
		sc.Table = &etable.Table{}
		sc.ConfigTable(sc.Table)
		yx := []string{"Y", "X"}
		sc.WorldTsr = etensor.NewFloat32([]int{sc.WorldVisSz.Y, sc.WorldVisSz.X}, nil, yx)
		sc.ViewTsr = etensor.NewFloat32([]int{sc.ViewVisSz.Y, sc.ViewVisSz.X}, nil, yx)
	}
}

func (sc *Saccade) ConfigTable(dt *etable.Table) {
	yx := []string{"Y", "X"}
	sch := etable.Schema{
		{"TrialName", etensor.STRING, nil, nil},
		{"Tick", etensor.INT64, nil, nil},
		{"SacTick", etensor.INT64, nil, nil},
		{"World", etensor.FLOAT32, []int{sc.WorldVisSz.Y, sc.WorldVisSz.X}, yx},
		{"View", etensor.FLOAT32, []int{sc.ViewVisSz.Y, sc.ViewVisSz.X}, yx},
		{"ObjPos", etensor.FLOAT32, []int{2}, nil},
		{"ObjViewPos", etensor.FLOAT32, []int{2}, nil},
		{"ObjVel", etensor.FLOAT32, []int{2}, nil},
		{"ObjPosNext", etensor.FLOAT32, []int{2}, nil},
		{"EyePos", etensor.FLOAT32, []int{2}, nil},
		{"SacPlan", etensor.FLOAT32, []int{2}, nil},
		{"Saccade", etensor.FLOAT32, []int{2}, nil},
	}
	dt.SetFromSchema(sch, 0)
}

func (sc *Saccade) WriteToTable(dt *etable.Table) {
	row := 0
	if sc.AddRows {
		row = dt.Rows
	}
	dt.SetNumRows(row + 1)
	dt.SetCellString("TrialName", row, "todo")
	dt.SetCellFloat("Tick", row, float64(sc.Tick.Cur))
	dt.SetCellFloat("SacTick", row, float64(sc.SacTick.Cur))

	sc.WorldTsr.SetZeros()
	opx := int(math.Round(float64(sc.ObjPos.X * float32(sc.WorldVisSz.X))))
	opy := int(math.Round(float64(sc.ObjPos.Y * float32(sc.WorldVisSz.Y))))
	sc.WorldTsr.SetFloat([]int{opy, opx}, 1)

	sc.ViewTsr.SetZeros()
	opx = int(math.Round(float64(sc.ObjViewPos.X * float32(sc.ViewVisSz.X))))
	opy = int(math.Round(float64(sc.ObjViewPos.Y * float32(sc.ViewVisSz.Y))))
	sc.ViewTsr.SetFloat([]int{opy, opx}, 1)

	dt.SetCellTensor("World", row, sc.WorldTsr)
	dt.SetCellTensor("View", row, sc.ViewTsr)

	dt.SetCellTensorFloat1D("ObjPos", row, 0, float64(sc.ObjPos.X))
	dt.SetCellTensorFloat1D("ObjPos", row, 1, float64(sc.ObjPos.Y))
	dt.SetCellTensorFloat1D("ObjViewPos", row, 0, float64(sc.ObjViewPos.X))
	dt.SetCellTensorFloat1D("ObjViewPos", row, 1, float64(sc.ObjViewPos.Y))
	dt.SetCellTensorFloat1D("ObjVel", row, 0, float64(sc.ObjVel.X))
	dt.SetCellTensorFloat1D("ObjVel", row, 1, float64(sc.ObjVel.Y))
	dt.SetCellTensorFloat1D("ObjPosNext", row, 0, float64(sc.ObjPosNext.X))
	dt.SetCellTensorFloat1D("ObjPosNext", row, 1, float64(sc.ObjPosNext.Y))
	dt.SetCellTensorFloat1D("EyePos", row, 0, float64(sc.EyePos.X))
	dt.SetCellTensorFloat1D("EyePos", row, 1, float64(sc.EyePos.Y))
	dt.SetCellTensorFloat1D("SacPlan", row, 0, float64(sc.SacPlan.X))
	dt.SetCellTensorFloat1D("SacPlan", row, 1, float64(sc.SacPlan.Y))
	dt.SetCellTensorFloat1D("Saccade", row, 0, float64(sc.Saccade.X))
	dt.SetCellTensorFloat1D("Saccade", row, 1, float64(sc.Saccade.Y))
}

func (sc *Saccade) ConstrainVelWorld(vel, start, trials float32) float32 {
	if trials <= 0 {
		return vel
	}
	end := start + vel*trials
	if end > sc.World.Max {
		vel = (sc.World.Max - start) / trials
	} else if end < sc.World.Min {
		vel = (sc.World.Min - start) / trials
	}
	return vel
}

func (sc *Saccade) ConstrainPos(pos, max float32) float32 {
	if pos > max {
		pos = max
	}
	if pos < -max {
		pos = -max
	}
	return pos
}

func (sc *Saccade) ConstrainSac(sacDev, start, objPos, objVel, trials float32) float32 {
	objEnd := objPos + objVel*trials
	eyep := start + sacDev
	lowView := eyep + sc.View.Min
	highView := eyep + sc.View.Max
	// do obj_end first then pos so it has stronger constraint
	if objEnd < lowView {
		sacDev += (objEnd - lowView)
	} else if objEnd > highView {
		sacDev += objEnd - highView
	}
	eyep = start + sacDev
	if eyep < sc.World.Min {
		sacDev += sc.World.Min - eyep
	} else if eyep > sc.World.Max {
		sacDev += sc.World.Max - eyep
	}
	eyep = start + sacDev
	lowView = eyep + sc.View.Min
	highView = eyep + sc.View.Max
	if objPos < lowView {
		sacDev += objPos - lowView
	} else if objPos > highView {
		sacDev += objPos - highView
	}
	eyep = start + sacDev
	if eyep < sc.World.Min {
		sacDev += sc.World.Min - eyep
	} else if eyep > sc.World.Max {
		sacDev += sc.World.Max - eyep
	}
	return sacDev
}

/*
int GenNextTraj() {
// local variables
bool init; init = 0;
init = false;
if(tick < 0 || saccade_input.rows == 0) {
InitTables();
traj_len = Random::IntMinMax(traj_len_min, traj_len_max+1);
init = true;
tick = 0;
{ PrintExpr* pexp = this.functions[4].fun_code[2].true_code[4];
if(!pexp->debug || pexp->InDebugMode()) {
cout << "\n\n ==============================" << endl;
}
}
}
if(!(env_opts & OBJ_MOVE)) {
next_obj_vel_x = 0;
next_obj_vel_y = 0;
}
else {
next_obj_vel_x = Random::UniformMeanRange(0, sc.View.Max);
next_obj_vel_y = Random::UniformMeanRange(0, sc.View.Max);
}
next_obj_pos_x = obj_pos_x + next_obj_vel_x;
next_obj_pos_y = obj_pos_y + next_obj_vel_y;
ConstrainPos(next_obj_pos_x, sc.World.Max);
ConstrainPos(next_obj_pos_y, sc.World.Max);
if(env_opts & OBJ_MOVE) {
if(env_opts & ZERO_VEL_OK) {
if(Random::BoolProb(p_zero_vel)) {
next_obj_vel_x = 0;
next_obj_vel_y = 0;
}
else {
next_obj_vel_x = Random::UniformMeanRange(0, vel_gen_max);
next_obj_vel_y = Random::UniformMeanRange(0, vel_gen_max);
ConstrainVelWorld(next_obj_vel_x, next_obj_pos_x, traj_len);
ConstrainVelWorld(next_obj_vel_y, next_obj_pos_y, traj_len);
}
}
else {
next_obj_vel_x = Random::UniformMeanRange(0, vel_gen_max);
next_obj_vel_y = Random::UniformMeanRange(0, vel_gen_max);
ConstrainVelWorld(next_obj_vel_x, next_obj_pos_x, traj_len);
ConstrainVelWorld(next_obj_vel_y, next_obj_pos_y, traj_len);
}
}
else {
next_obj_vel_x = 0;
next_obj_vel_y = 0;
}
{ PrintVar* pvar = this.functions[4].fun_code[11];
if(!pvar->debug || pvar->InDebugMode()) {
cout << "gen traj"<< " next_obj_pos_x = " << next_obj_pos_x<< " next_obj_pos_y = " << next_obj_pos_y<< " next_obj_vel_x = " << next_obj_vel_x<< " next_obj_vel_y = " << next_obj_vel_y<< " tick = " << tick<< " traj_len = " << traj_len << endl;
}
}
// always need new saccade at onset of new traj
GenNextSaccade();
if(init) {
traj_len = Random::IntMinMax(traj_len_min, traj_len_max+1);
NextToCur();
DoObjMove();
// for first one, move eyes now..
DoSaccade();
new_traj = true;
}
else {
// next trial is a new trajectory -- do it then
next_new_traj = true;
}
}
int GenNextSaccade() {
// local variables
int eff_trl; eff_trl = 0;
if(env_opts & SACCADE) {
saccade_plan_x = Random::UniformMeanRange(0, saccade_gen_max);
saccade_plan_y = Random::UniformMeanRange(0, saccade_gen_max);
}
else {
saccade_plan_x = 0;
saccade_plan_y = 0;
}
fix_dur = Random::IntMinMax(fix_dur_min, fix_dur_max+1);
eff_trl = fix_dur;
if(!(env_opts & SACCADE)) {
eff_trl = traj_len;
}
ConstrainSaccade(saccade_plan_x, eye_pos_x, next_obj_pos_x, next_obj_vel_x, eff_trl);
ConstrainSaccade(saccade_plan_y, eye_pos_y, next_obj_pos_y, next_obj_vel_y, eff_trl);
{ PrintVar* pvar = this.functions[5].fun_code[9];
if(!pvar->debug || pvar->InDebugMode()) {
cout << "gen saccade"<< " eye_pos_x = " << eye_pos_x<< " eye_pos_y = " << eye_pos_y<< " saccade_plan_x = " << saccade_plan_x<< " saccade_plan_y = " << saccade_plan_y<< " saccade_tick = " << saccade_tick<< " fix_dur = " << fix_dur << endl;
}
}
}
int DoSaccade() {
{ PrintVar* pvar = this.functions[6].fun_code[1];
if(!pvar->debug || pvar->InDebugMode()) {
cout << "do saccade"<< " eye_pos_x = " << eye_pos_x<< " eye_pos_y = " << eye_pos_y<< " saccade_plan_x = " << saccade_plan_x<< " saccade_plan_y = " << saccade_plan_y<< " saccade_tick = " << saccade_tick << endl;
}
}
eye_pos_x = eye_pos_x + saccade_plan_x;
eye_pos_y = eye_pos_y+ saccade_plan_y;
saccade_x = saccade_plan_x;
saccade_y = saccade_plan_y;
did_saccade = true;
saccade_tick = 0;
saccade_plan_x = 0;
saccade_plan_y = 0;
}
int DoneSaccade() {
saccade_x = 0;
saccade_y = 0;
did_saccade = false;
}
int PlanObjMove() {
if(next_new_traj) {
tick = 0;
new_traj = true;
next_new_traj = false;
{ PrintVar* pvar = this.functions[8].fun_code[1].true_code[3];
if(!pvar->debug || pvar->InDebugMode()) {
cout << "\n=== start new traj "<< " next_obj_pos_x = " << next_obj_pos_x<< " next_obj_pos_y = " << next_obj_pos_y<< " obj_pos_x = " << obj_pos_x<< " obj_pos_y = " << obj_pos_y<< " obj_vel_x = " << obj_vel_x<< " obj_vel_y = " << obj_vel_y << endl;
}
}
}
next_obj_pos_x = obj_pos_x + obj_vel_x;
next_obj_pos_y = obj_pos_y + obj_vel_y;
{ PrintVar* pvar = this.functions[8].fun_code[4];
if(!pvar->debug || pvar->InDebugMode()) {
cout << "plan move"<< " next_obj_pos_x = " << next_obj_pos_x<< " next_obj_pos_y = " << next_obj_pos_y<< " obj_pos_x = " << obj_pos_x<< " obj_pos_y = " << obj_pos_y<< " obj_vel_x = " << obj_vel_x<< " obj_vel_y = " << obj_vel_y << endl;
}
}
if((env_opts & SACCADE) && (saccade_tick+1 >= fix_dur)) {
GenNextSaccade();
}
}
int DoObjMove() {
{ PrintVar* pvar = this.functions[9].fun_code[1];
if(!pvar->debug || pvar->InDebugMode()) {
cout << "move obj"<< " obj_pos_x = " << obj_pos_x<< " obj_pos_y = " << obj_pos_y<< " next_obj_pos_x = " << next_obj_pos_x<< " next_obj_pos_y = " << next_obj_pos_y << endl;
}
}
obj_pos_x = next_obj_pos_x;
obj_pos_y = next_obj_pos_y;
}
int NextToCur() {
if(next_new_traj) {
obj_vel_x = next_obj_vel_x;
obj_vel_y = next_obj_vel_y;
}
}
int RenderXYCol(String xy_col_nm, double x, double y) {
saccade_input[xy_col_nm][0,0,-1] = x;
saccade_input[xy_col_nm][1,0,-1] = y;
}
int RangeCheckErr(double& val, double maxval, String valnm) {
if(val - maxval > compare_tol) {
{ PrintVar* pvar = this.functions[12].fun_code[1].true_code[0];
if(!pvar->debug || pvar->InDebugMode()) {
cout << "error value above range"<< " val = " << val<< " maxval = " << maxval<< " valnm = " << valnm << endl;
}
}
val = maxval;
}
if(val - -maxval < -compare_tol) {
{ PrintVar* pvar = this.functions[12].fun_code[2].true_code[0];
if(!pvar->debug || pvar->InDebugMode()) {
cout << "error value below range"<< " val = " << val<< " maxval = " << maxval<< " valnm = " << valnm << endl;
}
}
val = -maxval;
}
}
int GenInput() {
// local variables
String gp_name; gp_name = "";
DataCol* col; col = NULL;
int world_x; world_x = 0;
int world_y; world_y = 0;
int view_x; view_x = 0;
int view_y; view_y = 0;
int meta_idx; meta_idx = 0;
new_traj = false;
// needs init
if(tick < 0 || tick + 1 == traj_len) {
GenNextTraj();
}
else {
PlanObjMove();
}
if(add_rows) {
saccade_input->AddBlankRow();
}
else {
saccade_input->EnforceRows(1);
if(disp_opts & CLEAR_STEP) {
foreach(col in saccade_input.data) {
col->InitVals(0);
}
}
}
if(tick == 0 && disp_opts & CLEAR_TRAJ) {
foreach(col in saccade_input.data) {
col->InitVals(0);
}
}
saccade_input->ReadItem(-1);
obj_pos_x_view = obj_pos_x - eye_pos_x;
obj_pos_y_view = obj_pos_y - eye_pos_y;
RenderXYCol("obj_pos_xy", obj_pos_x, obj_pos_y);
RenderXYCol("obj_pos_xy_view", obj_pos_x_view, obj_pos_y_view);
RenderXYCol("obj_next_pos_xy", next_obj_pos_x, next_obj_pos_y);
RenderXYCol("obj_vel_xy", obj_vel_x, obj_vel_y);
RenderXYCol("eye_pos_xy", eye_pos_x, eye_pos_y);
RenderXYCol("saccade_plan_xy", saccade_plan_x, saccade_plan_y);
RenderXYCol("saccade_xy", saccade_x, saccade_y);
gp_name = "x_" + obj_pos_x + "_y_" +obj_pos_y;
saccade_input["group"][-1] = gp_name;
saccade_input["name"][-1] = "tick_" + String(tick) + "_sac_" + String(saccade_tick);
saccade_input->SetMatrixFlatVal(1, "what_id", -1, cur_category_id);
meta_idx = meta_categs_lba->FindVal(cur_category, "category", 0, true);
meta_id = meta_categs_lba["meta_lba_id"][meta_idx];
saccade_input->InitVals(0, "what_id_meta", 0, -1);
saccade_input->SetMatrixFlatVal(1, "what_id_meta", -1, meta_id);
if(taMisc::gui_active) {
RangeCheckErr(obj_pos_x, sc.World.Max, "obj_pos_x");
RangeCheckErr(obj_pos_y, sc.World.Max, "obj_pos_y");
RangeCheckErr(obj_pos_x_view, sc.View.Max, "obj_pos_x_view");
RangeCheckErr(obj_pos_y_view, sc.View.Max, "obj_pos_y_view");
world_x = (0.5 * (obj_pos_x + 1.0)) * world_width;
world_y = (0.5 * (obj_pos_y + 1.0)) * world_height;
view_x = ((0.5 * (obj_pos_x_view + view_pct)) / view_pct) * view_width;
view_y = ((0.5 * (obj_pos_y_view + view_pct)) / view_pct) * view_height;
saccade_input["world"][world_x, world_y, -1] = 1.0;
saccade_input["view"][view_x, view_y, -1] = 1.0;
}
{ PrintVar* pvar = this.functions[13].fun_code[27];
if(!pvar->debug || pvar->InDebugMode()) {
cout << "--- rendered "<< " tick = " << tick<< " saccade_tick = " << saccade_tick << endl;
}
}
saccade_input->UpdateAllViews();
NextToCur();
DoObjMove();
if(next_new_traj || saccade_tick+1 >= fix_dur) {
DoSaccade();
}
else {
if(did_saccade) {
DoneSaccade();
}
saccade_tick = saccade_tick + 1;
}
tick = tick + 1;
}
void __Init() {
// init_from vars
{ // init_from
Program* init_fm_prog = this.vars[46]->GetInitFromProg();
cur_category = init_fm_prog->GetVar("cur_category");
}
{ // init_from
Program* init_fm_prog = this.vars[47]->GetInitFromProg();
cur_category_id = init_fm_prog->GetVar("cur_category_id");
}
// run our init code
new_traj = false;
next_new_traj = false;
traj_len = 0;
eye_pos_x = 0;
eye_pos_y = 0;
obj_pos_x = 0;
obj_pos_y = 0;
saccade_plan_x = 0;
saccade_plan_y = 0;
saccade_x = 0;
saccade_y = 0;
obj_vel_x = 0;
obj_vel_y = 0;
next_obj_vel_x = 0;
next_obj_vel_y = 0;
tick = -1;
saccade_tick = 0;
}

*/
