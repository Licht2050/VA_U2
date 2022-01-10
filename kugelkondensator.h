#ifndef KUGELKONDENSATOR
#define KUGELKONDENSATOR

struct Kondensator
{
    double inRadius;
    double ausRadius;
};

const double PI = 3.14159;
const double EPSILON_NULL = 8.8543E-12 * 1.00059;

bool userInput (Kondensator & k);
double berechnung(const Kondensator & k);




#endif