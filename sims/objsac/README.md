# ObjSac

This does predictive learning of saccade-related signals *only*, in contrast to the wwi3d model which predicts the shape of the object as well.  Thus, this model should be able to achieve essentially perfect predictive accuracy.

# Critical Bugs / Issues

The start of a new trajectory resets the position of the object, in a way that is *hidden* from the model -- the relationship between the eye position and view is dissociated.  In principle, it could use a working memory-like representation to maintain the current world position of the object, but we're not giving the model that opportunity.

This issue was obscured in the main model because just predicting the object movement and shape, plus some of the eye movement update, was generally "good enough".


