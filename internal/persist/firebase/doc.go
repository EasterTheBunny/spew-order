package firebase

/*
	Package firebase contains repository implementations for the Google firestore
	data store. One of the notable functions of this repository is the extra code
	added to handle write contention. Google firestore suggests not writing to the
	same object more than once per second. Each implemented method handles its own
	write contention cases and should be safe to use on its own.
*/
