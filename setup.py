import setuptools

# from https://gist.github.com/dcreager/300803 with "-dirty" support added
from version import get_git_version

# From http://bugs.python.org/issue15881
try:
    import multiprocessing
except ImportError:
    pass

with open("README.md", "r") as fh:
    long_description = fh.read()

setuptools.setup(
    name="gowerline",
    version=get_git_version(),
    author="Thomas Maurice",
    author_email="thomas@maurice.fr",
    description="Write your powerline segments in Go !",
    long_description=long_description,
    long_description_content_type='text/markdown',
    url='https://github.com/thomas-maurice/gowerline',
    packages=setuptools.find_packages(),
    package_dir={"": "."},
    install_requires=[
        'powerline-status',
    ],
    classifiers=[
        'Development Status :: 4 - Beta',
        'Environment :: Console',
        'Intended Audience :: Developers',
        'Intended Audience :: System Administrators',
        'Intended Audience :: End Users/Desktop',
        'License :: OSI Approved :: MIT License',
        'Operating System :: POSIX',
        'Programming Language :: Python :: 3.7',
        'Programming Language :: Python :: 3 :: Only',
        'Topic :: Utilities',
        'Topic :: System :: Shells',
    ],
)
