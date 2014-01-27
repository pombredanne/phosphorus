%module cephes
%{
extern double incbet(double aa, double bb, double xx);
extern double beta(double a, double b);
%}

extern double incbet(double aa, double bb, double xx);
extern double beta(double a, double b);