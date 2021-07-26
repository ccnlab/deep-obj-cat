# Saccade

This does predictive learning of saccade-related signals *only*, in contrast to the wwi3d model which predicts the shape of the object as well.  Thus, this model should be able to achieve essentially perfect predictive accuracy.

# Multiple primary predicted signals

The primary signals to predict are:

* `V1` visual image (idealized blob), on a 2D retinotopic map

* `S1e` primary proprioceptive somatosensory eye position map (WangZhangCohenEtAl07) -- cells have gaussian direction and graded slope eccentricity coding, roughly linear from the center out.  Orientation is coded on X axis with vertical in center, angles to the left on the left, etc.  To maintain consistency with FEF motor code, and have a simpler overall activity level profile, eccentricity coding is also gaussian with a preferred eccentricity coding progressively outward across the higher rows.

* `FEF` has a topographically organized map in polar coordinates (orientation and eccentricity) (SommerWurtz00), organized in the model like S1e.

`LIP` receives from and predicts all of the above primary signals.

# Paradigm

Ideally: multiple blobs in the input, fixate on an attentionally-selected one in the periphery, predict where the attentional blob (and others?) go.

WWI3D model: saccading around an object that can be moving too.

Scaffold: single blob, fixate on it.

## Timeline: 

* T0: visual input presented, saccade planned (FEF super) then executed in plus phase (-> FEF deep).  can have random initial eye position.

* T1: new visual input, based on actual saccade, predicted vs. actual for all things like eye position etc
