# Saccade

This does predictive learning of saccade-related signals *only*, in contrast to the wwi3d model which predicts the shape of the object as well.  Thus, this model should be able to achieve essentially perfect predictive accuracy.

# Network Layers

The primary layers in the model are:

* `V1p` primary visual cortex, peripheral, full-field, represents the visual image (idealized blobs), on a 2D retinotopic map: one or more blobs, one of which is the target.

* `S1e` primary proprioceptive somatosensory eye position map (WangZhangCohenEtAl07) -- cells have gaussian direction and graded slope eccentricity coding, roughly linear from the center out.  Orientation is coded on X axis with vertical in center, angles to the left on the left, etc.  To maintain consistency with FEF motor code, and have a simpler overall activity level profile, eccentricity coding is also gaussian with a preferred eccentricity coding progressively outward across the higher rows.

* `SCd` is the deep superior colliculus, which is the motor output for eye movement.  It has a topographically organized map in polar coordinates (orientation and eccentricity) organized in the model like S1e.  

* `SCs` is the superficial superior colliculus, which generates saccade plans based on sensory inputs.  It has a topographically organized map in polar coordinates (orientation and eccentricity) organized in the model like S1e and SCd.  

* `MDe` is the medial dorsal thalamus, representing the eye motor signals.  It receives an corollary discharge projection from the SCd, reflecting the actual eye motor action taken, driving the plus phase on the MD.  MD also receives a top-down projection from FEF, representing a kind of prediction relative to the SCd plus phase -- this error signal is how the FEF learns to send signals in the proper language of the motor system.

* `FEF` is the primary motor frontal eye fields, sending activity to SCd to drive cortically-driven eye movements, and also a corollary signal to MDe representing the prediction.

* `SEF` is supplementary / second-order eye fields, which learns to predict FEF activity, and provides more strategic top-down motor plans to FEF, e.g., for sequencing etc.

* `LIP` receives from and predicts V1, S1e, and FEF, sending activity to FEF and SEF to drive saccades to the attentionally-selected V1 target.

## Circuits

* `V1 -> LIP <-> FEF <-> MDe -> SCd` -- basic sensory-motor pathway, where the visual target represented in LIP drives motor output, with FEF tansforming the retinotopic coordinates of LIP into the polar motor coordinates of MD which is constrained by SCd for its representational structure.  In the model, the MDe plus phase driven by SCd corollary discharge acts like a standard target layer.

# Paradigm

Multiple blobs in the input, fixate on an attentionally-selected one in the periphery, predict where the attentional blob (and others?) go.

## Timeline: 

* T0: visual input presented, saccade planned (FEF super) then executed in plus phase (-> FEF deep).  can have random initial eye position.

* T1: new visual input, based on actual saccade, predicted vs. actual for all things like eye position etc
